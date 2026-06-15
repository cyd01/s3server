package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SigV4Authenticator struct {
	credentials CredentialStore
	region      string
	service     string
}

func NewSigV4Authenticator(
	credentials CredentialStore,
	region string,
	service string,
) *SigV4Authenticator {
	return &SigV4Authenticator{
		credentials: credentials,
		region:      region,
		service:     service,
	}
}

func (a *SigV4Authenticator) Authenticate(r *http.Request) error {
	auth := r.Header.Get("Authorization")

	if auth == "" {
		return errors.New("missing authorization header")
	}

	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256 ") {
		return errors.New("unsupported authorization scheme")
	}

	parsed, err := parseAuthorization(auth)
	if err != nil {
		return err
	}

	secret, ok := a.credentials.SecretKey(parsed.AccessKeyID)
	if !ok {
		return errors.New("unknown access key")
	}

	amzDate := r.Header.Get("X-Amz-Date")
	if amzDate == "" {
		return errors.New("missing X-Amz-Date")
	}

	t, err := time.Parse("20060102T150405Z", amzDate)
	if err != nil {
		return errors.New("invalid X-Amz-Date")
	}

	if d := time.Since(t); d > 15*time.Minute || d < -15*time.Minute {
		return errors.New("request time too skewed")
	}

	canonicalRequest, err := buildCanonicalRequest(r, parsed.SignedHeaders)
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(canonicalRequest))

	credentialScope := fmt.Sprintf(
		"%s/%s/%s/aws4_request",
		parsed.Date,
		parsed.Region,
		parsed.Service,
	)

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hex.EncodeToString(hash[:]),
	}, "\n")

	signingKey := deriveSigningKey(
		secret,
		parsed.Date,
		parsed.Region,
		parsed.Service,
	)

	mac := hmac.New(sha256.New, signingKey)
	mac.Write([]byte(stringToSign))

	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(parsed.Signature)) {
		return errors.New("signature mismatch")
	}

	return nil
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func deriveSigningKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}
