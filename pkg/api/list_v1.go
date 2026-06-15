package api

import (
	"encoding/xml"
	"net/http"
)

func (h *Handler) listObjectsV1(w http.ResponseWriter, r *http.Request, bucket string) {
	objects, err := h.storage.ListObjects(r.Context(), bucket)
	if err != nil {
		WriteError(w, "NoSuchBucket", err.Error(), 404)
		return
	}

	resp := ListBucketResult{
		Name:     bucket,
		KeyCount: len(objects),
	}

	for _, o := range objects {
		resp.Contents = append(resp.Contents, Object{
			Key:          o.Key,
			Size:         o.Size,
			LastModified: o.LastModified,
		})
	}

	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(resp)

}
