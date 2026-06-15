package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const maxDebugLen = 4096

func Log(debug bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Method, r.Host, r.URL.Path, r.ContentLength)
		if debug {
			for k, v := range r.Header {
				fmt.Println("  ", k, ":", strings.Join(v, ","))
			}
			fmt.Println("")
			if (r.ContentLength > 0) && (r.ContentLength <= maxDebugLen) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Unable to read request body", http.StatusBadRequest)
					return
				}
				_ = r.Body.Close()
				fmt.Println("  ", string(body))
				r.Body = io.NopCloser(bytes.NewReader(body))
			}
		}
		next.ServeHTTP(w, r)
	})
}
