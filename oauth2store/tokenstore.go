package oauth2store

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"codeberg.org/clambin/go-crypt"
	"golang.org/x/oauth2"
)

// Storer interface decouples the TokenStore from the underlying storage mechanism.
type Storer interface {
	Save([]byte) error
	Load() ([]byte, error)
}

// A TokenStore saves and loads oauth2 Tokens. In conjunction with this package's TokenSource,
// it allows calling applications to persist tokens and reuse them between runs.
//
// The main use case is to persist tokens that require manual action to create (e.g. device authentication flow), so that
// a previously created token may be reused if the application is restarted.
type TokenStore struct {
	Path string
	Storer
}

// NewEncryptedFileTokenStore returns a TokenStore that stored the encrypted token to disk.
// Expired tokens are not loaded.
func NewEncryptedFileTokenStore(path, passphrase string) TokenStore {
	return TokenStore{
		Path:   path,
		Storer: crypt.New(path, passphrase),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t TokenStore) Save(token *oauth2.Token) error {
	bytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}
	return t.Storer.Save(bytes)
}

func (t TokenStore) Load() (*oauth2.Token, error) {
	bytes, err := t.Storer.Load()
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	if err = json.Unmarshal(bytes, &token); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	return &token, nil
}

func (t TokenStore) LastSaved() time.Time {
	if fileInfo, err := os.Stat(t.Path); err == nil {
		return fileInfo.ModTime()
	}
	return time.Time{}
}
