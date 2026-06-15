package auth

import (
	"context"
	"encoding/xml"
	"net/http"
)

type Middleware struct {
	authenticator Authenticator
}

func NewMiddleware(a Authenticator) *Middleware {
	return &Middleware{
		authenticator: a,
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if err := m.authenticator.Authenticate(r); err != nil {
			writeError(
				w,
				"AccessDenied",
				err.Error(),
				http.StatusForbidden,
			)
			return
		}

		// Optionnel : enrichir le contexte avec l'AccessKeyID.
		if extractor, ok := m.authenticator.(interface {
			AccessKeyID(*http.Request) string
		}); ok {

			if accessKey := extractor.AccessKeyID(r); accessKey != "" {
				ctx := context.WithValue(
					r.Context(),
					ContextAccessKeyID,
					accessKey,
				)

				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

type errorResponse struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	_ = xml.NewEncoder(w).Encode(errorResponse{
		Code:    code,
		Message: message,
	})
}
