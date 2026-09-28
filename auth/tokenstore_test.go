package auth

import (
	"errors"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func TestEncryptedFileStore(t *testing.T) {
	const goodPassphrase = "good-passphrase"

	// empty store. no token
	path := filepath.Join(t.TempDir(), "token.enc")
	store := NewEncryptedFileTokenStore(path, goodPassphrase)
	_, err := store.Load()
	if !errors.Is(err, ErrNoTokenFound) {
		t.Fatalf("Load() error = %v, wantErr %v", err, ErrNoTokenFound)
	}

	// save a token
	if err = store.Save(&oauth2.Token{AccessToken: "valid-token"}); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	// load token
	token, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load token: %v", err)
	}
	if got := token.AccessToken; got != "valid-token" {
		t.Fatalf("Token() = %v, want valid-token", got)
	}

	// invalid passphrase fails decryption
	store = NewEncryptedFileTokenStore(path, "invalid-passphrase")
	_, err = store.Load()
	if got := err.Error(); got != "crypt: invalid key: cipher: message authentication failed" {
		t.Fatalf("Load() error = %v", err)
	}
}
