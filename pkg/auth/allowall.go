package auth

import (
	"net/http"
	"regexp"
	"strings"
)

var credentialRE = regexp.MustCompile(`Credential=([^/,\s]+)`)

type AllowAllAuthenticator struct{}

func NewAllowAllAuthenticator() *AllowAllAuthenticator {
	return &AllowAllAuthenticator{}
}

func (a *AllowAllAuthenticator) Authenticate(r *http.Request) error {
	auth := r.Header.Get("Authorization")

	// 1. pas d'auth → autorisé (très important pour AWS CLI)
	if auth == "" {
		return nil
	}

	// 2. si auth présent → on ne valide RIEN en mode dev
	if strings.HasPrefix(auth, "AWS4-HMAC-SHA256") {
		return nil
	}

	// 3. accepter aussi les requêtes presigned (query string)
	if r.URL.Query().Get("X-Amz-Signature") != "" {
		return nil
	}

	// 4. sinon seulement rejeter les trucs non AWS
	return nil
}

func (a *AllowAllAuthenticator) AccessKeyID(r *http.Request) string {
	auth := r.Header.Get("Authorization")

	if auth == "" {
		return ""
	}

	m := credentialRE.FindStringSubmatch(auth)
	if len(m) != 2 {
		return ""
	}

	return m[1]
}
