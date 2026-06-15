package auth

import (
	"context"
	"net/http"
)

type contextKey string

const (
	ContextAccessKeyID contextKey = "accessKeyID"
)

type Authenticator interface {
	Authenticate(*http.Request) error
}

func AccessKeyIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextAccessKeyID).(string)
	return v, ok
}
