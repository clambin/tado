package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"codeberg.org/clambin/go-crypt"
	"golang.org/x/oauth2"
)

var ErrNoTokenFound = errors.New("no token found")

type TokenStore interface {
	Load() (*oauth2.Token, error)
	Save(*oauth2.Token) error
}

// EncryptedFileTokenStore is a [TokenStore] that saves the token to a file,
// using AES-256 encryption and a (salted) passphrase.
type EncryptedFileTokenStore struct {
	crypt.Crypt
}

// NewEncryptedFileTokenStore returns a TokenStore that stored the encrypted token to disk.
// Expired tokens are not loaded.
func NewEncryptedFileTokenStore(path, passphrase string) *EncryptedFileTokenStore {
	return &EncryptedFileTokenStore{Crypt: crypt.New(path, passphrase)}
}
func (s *EncryptedFileTokenStore) Save(token *oauth2.Token) error {
	bytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}
	return s.Crypt.Save(bytes)
}

func (s *EncryptedFileTokenStore) Load() (*oauth2.Token, error) {
	bytes, err := s.Crypt.Load()
	if err == nil {
		var token oauth2.Token
		if err = json.Unmarshal(bytes, &token); err != nil {
			return nil, fmt.Errorf("json: %w", err)
		}
		return &token, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoTokenFound
	}
	return nil, fmt.Errorf("crypt: %w", err)
}
