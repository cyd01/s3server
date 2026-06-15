package api

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cyd01/s3server/pkg/storage"
)

type Handler struct {
	storage storage.Storage
}

func NewHandler(s storage.Storage) *Handler {
	return &Handler{storage: s}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if r.URL.Path == "/" && r.Method == http.MethodGet {
		h.listBuckets(w, r)
		return
	}

	if len(parts) == 1 {
		h.handleBucket(w, r, parts[0])
		return
	}

	if len(parts) >= 2 {
		key := strings.Join(parts[1:], "/")
		h.handleObject(w, r, parts[0], key)
		return
	}

	WriteError(w, "NotImplemented", "operation not supported", http.StatusNotImplemented)
}

func (h *Handler) listBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.storage.ListBuckets(context.Background())
	if err != nil {
		WriteError(w, "InternalError", err.Error(), 500)
		return
	}

	resp := ListBucketsResult{}

	for _, b := range buckets {
		resp.Buckets = append(resp.Buckets, Bucket{
			Name:         b,
			CreationDate: time.Now().UTC(),
		})
	}

	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleBucket(w http.ResponseWriter, r *http.Request, bucket string) {

	switch r.Method {

	case http.MethodPut:
		if err := h.storage.CreateBucket(r.Context(), bucket); err != nil {
			WriteError(w, "BucketAlreadyExists", err.Error(), 409)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		_ = h.storage.DeleteBucket(r.Context(), bucket)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodHead:
		ok, _ := h.storage.BucketExists(r.Context(), bucket)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodGet:
		if r.URL.Query().Get("list-type") == "2" {
			h.listObjectsV2(w, r, bucket)
			return
		} else {
			h.listObjectsV1(w, r, bucket)
			return
		}

		//WriteError(w, "NotImplemented", "only ListObjectsV2 supported", 501)
	}
}

func (h *Handler) handleObject(w http.ResponseWriter, r *http.Request, bucket, key string) {
	switch r.Method {

	case http.MethodPut:
		if err := h.storage.PutObject(r.Context(), bucket, key, r.Body); err != nil {
			WriteError(w, "InternalError", err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusOK)

	case http.MethodGet:

		rc, info, err := h.storage.GetObject(r.Context(), bucket, key)
		if err != nil {
			WriteError(w, "NoSuchKey", err.Error(), 404)
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", info.ContentType)

		// ⭐ ETag réel
		if fs, ok := h.storage.(*storage.FileSystem); ok {
			etag := fs.GetETag(fs.ObjectPath(bucket, key))
			if etag != "" {
				w.Header().Set("ETag", etag)
			}
		}

		// ⭐ Range support (CRITIQUE pour aws s3 cp / resume)
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			start, end := parseRange(rangeHeader, info.Size)

			w.Header().Set("Content-Range",
				fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size))

			w.WriteHeader(http.StatusPartialContent)

			_, _ = io.CopyN(w, rc, end-start+1)
			return
		}

		w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
		_, _ = io.Copy(w, rc)

	case http.MethodHead:

		info, err := h.storage.HeadObject(r.Context(), bucket, key)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
		w.Header().Set("Content-Type", info.ContentType)

		if fs, ok := h.storage.(*storage.FileSystem); ok {
			w.Header().Set("ETag", fs.GetETag(fs.ObjectPath(bucket, key)))
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		if err := h.storage.DeleteObject(r.Context(), bucket, key); err != nil {
			WriteError(w, "NoSuchKey", err.Error(), 404)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func parseRange(h string, size int64) (int64, int64) {
	// format: bytes=start-end
	if !strings.HasPrefix(h, "bytes=") {
		return 0, size - 1
	}

	r := strings.TrimPrefix(h, "bytes=")
	parts := strings.Split(r, "-")

	start := int64(0)
	end := size - 1

	if parts[0] != "" {
		start, _ = strconv.ParseInt(parts[0], 10, 64)
	}

	if len(parts) > 1 && parts[1] != "" {
		end, _ = strconv.ParseInt(parts[1], 10, 64)
	}

	if end >= size {
		end = size - 1
	}

	return start, end
}
