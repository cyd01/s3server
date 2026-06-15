package api

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"time"
)

type ListBucketV2Response struct {
	XMLName               xml.Name `xml:"ListBucketResult"`
	Name                  string   `xml:"Name"`
	Prefix                string   `xml:"Prefix"`
	KeyCount              int      `xml:"KeyCount"`
	MaxKeys               int      `xml:"MaxKeys"`
	IsTruncated           bool     `xml:"IsTruncated"`
	NextContinuationToken string   `xml:"NextContinuationToken,omitempty"`

	Contents []ObjectV2 `xml:"Contents"`
}

type ObjectV2 struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

func (h *Handler) listObjectsV2(w http.ResponseWriter, r *http.Request, bucket string) {

	q := r.URL.Query()

	prefix := q.Get("prefix")

	maxKeys := 1000
	if v := q.Get("max-keys"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			maxKeys = i
		}
	}

	continuation := q.Get("continuation-token")

	objects, next, err := h.storage.ListObjectsV2(
		r.Context(),
		bucket,
		prefix,
		continuation,
		maxKeys,
	)
	if err != nil {
		WriteError(w, "NoSuchBucket", err.Error(), 404)
		return
	}

	resp := ListBucketV2Response{
		Name:        bucket,
		Prefix:      prefix,
		KeyCount:    len(objects),
		MaxKeys:     maxKeys,
		IsTruncated: next != "",
	}

	if next != "" {
		resp.NextContinuationToken = next
	}

	for _, o := range objects {
		resp.Contents = append(resp.Contents, ObjectV2{
			Key:          o.Key,
			LastModified: o.LastModified.UTC().Format(time.RFC3339),
			ETag:         o.ETag,
			Size:         o.Size,
			StorageClass: "STANDARD",
		})
	}

	w.Header().Set("Content-Type", "application/xml")
	_ = xml.NewEncoder(w).Encode(resp)
}
