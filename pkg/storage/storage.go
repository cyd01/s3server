package storage

import (
	"context"
	"io"
	"time"
)

type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
}

type Storage interface {
	ListBuckets(context.Context) ([]string, error)

	CreateBucket(context.Context, string) error
	DeleteBucket(context.Context, string) error
	BucketExists(context.Context, string) (bool, error)

	ListObjects(context.Context, string) ([]ObjectInfo, error)
	ListObjectsV2(ctx context.Context, bucket, prefix, continuation string, maxKeys int) ([]ObjectInfo, string, error)

	PutObject(context.Context, string, string, io.Reader) error

	GetObject(context.Context, string, string) (io.ReadCloser, ObjectInfo, error)

	HeadObject(context.Context, string, string) (ObjectInfo, error)

	DeleteObject(context.Context, string, string) error
}
