package auth

type authorization struct {
	AccessKeyID   string
	Date          string
	Region        string
	Service       string
	SignedHeaders []string
	Signature     string
}

func parseAuthorization(header string) (*authorization, error) {
	// À implémenter proprement :
	// Credential=AKIA.../20260617/us-east-1/s3/aws4_request
	// SignedHeaders=...
	// Signature=...

	return nil, nil
}
