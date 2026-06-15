package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem struct {
	root string
}

func NewFileSystem(root string) *FileSystem {
	return &FileSystem{root: root}
}

func (fs *FileSystem) bucketPath(bucket string) string {
	return filepath.Join(fs.root, bucket)
}

func (fs *FileSystem) ObjectPath(bucket, key string) string {
	return filepath.Join(fs.root, bucket, filepath.Clean(key))
}

func (fs *FileSystem) ListBuckets(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(fs.root)
	if err != nil {
		return nil, err
	}

	var buckets []string

	for _, e := range entries {
		if e.IsDir() {
			buckets = append(buckets, e.Name())
		}
	}

	return buckets, nil
}

func (fs *FileSystem) CreateBucket(ctx context.Context, bucket string) error {
	return os.MkdirAll(fs.bucketPath(bucket), 0755)
}

func (fs *FileSystem) DeleteBucket(ctx context.Context, bucket string) error {
	return os.Remove(fs.bucketPath(bucket))
}

func (fs *FileSystem) BucketExists(ctx context.Context, bucket string) (bool, error) {
	st, err := os.Stat(fs.bucketPath(bucket))

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return st.IsDir(), err
}

func (fs *FileSystem) ListObjects(ctx context.Context, bucket string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	root := fs.bucketPath(bucket)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if isInternalFile(info.Name()) {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		objects = append(objects, ObjectInfo{
			Key:          filepath.ToSlash(rel),
			Size:         info.Size(),
			LastModified: info.ModTime(),
			ContentType:  mime.TypeByExtension(filepath.Ext(path)),
		})

		return nil
	})

	return objects, err
}

func (fs *FileSystem) PutObject(ctx context.Context, bucket, key string, r io.Reader) error {

	path := fs.ObjectPath(bucket, key)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	hash := md5.New()
	writer := io.MultiWriter(f, hash)

	if _, err := io.Copy(writer, r); err != nil {
		f.Close()
		return err
	}

	f.Close()

	etag := hex.EncodeToString(hash.Sum(nil))

	// stocker ETag dans un sidecar file
	if err := os.WriteFile(tmp+".etag", []byte(etag), 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func (fs *FileSystem) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectInfo, error) {

	path := fs.ObjectPath(bucket, key)

	f, err := os.Open(path)
	if err != nil {
		return nil, ObjectInfo{}, err
	}

	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, ObjectInfo{}, err
	}

	info := ObjectInfo{
		Key:          key,
		Size:         st.Size(),
		LastModified: st.ModTime(),
	}

	return f, info, nil
}

func (fs *FileSystem) HeadObject(ctx context.Context, bucket, key string) (ObjectInfo, error) {
	path := fs.ObjectPath(bucket, key)

	st, err := os.Stat(path)
	if err != nil {
		return ObjectInfo{}, err
	}

	return ObjectInfo{
		Key:          key,
		Size:         st.Size(),
		LastModified: st.ModTime(),
		ContentType:  mime.TypeByExtension(filepath.Ext(path)),
	}, nil
}

func (fs *FileSystem) DeleteObject(ctx context.Context, bucket, key string) error {
	if os.Remove(fs.ObjectPath(bucket, key)+".etag") != nil {
		os.Remove(fs.ObjectPath(bucket, key) + ".tmp.etag")
	}
	return os.Remove(fs.ObjectPath(bucket, key))
}

func (fs *FileSystem) ListObjectsV2(
	ctx context.Context,
	bucket, prefix, continuation string,
	maxKeys int,
) ([]ObjectInfo, string, error) {

	root := fs.bucketPath(bucket)

	var out []ObjectInfo
	count := 0
	nextToken := ""

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if isInternalFile(info.Name()) {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		key := filepath.ToSlash(rel)

		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}

		if continuation != "" && key <= continuation {
			return nil
		}

		if count >= maxKeys {
			nextToken = key
			return filepath.SkipDir
		}

		etag := fs.GetETag(path)

		out = append(out, ObjectInfo{
			Key:          key,
			Size:         info.Size(),
			LastModified: info.ModTime(),
			ETag:         etag,
		})

		count++
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return out, nextToken, nil
}
