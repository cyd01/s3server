package auth

type CredentialStore interface {
	SecretKey(accessKeyID string) (string, bool)
}

type StaticCredentialStore struct {
	credentials map[string]string
}

func NewStaticCredentialStore(creds map[string]string) *StaticCredentialStore {
	return &StaticCredentialStore{
		credentials: creds,
	}
}

func (s *StaticCredentialStore) SecretKey(accessKeyID string) (string, bool) {
	v, ok := s.credentials[accessKeyID]
	return v, ok
}
