package oauth2store_test

import (
	"path/filepath"
	"testing"

	"github.com/clambin/tado/v2/oauth2store"
	"golang.org/x/oauth2"
)

func TestEncryptedFileStore_Load(t *testing.T) {
	tests := []struct {
		name       string
		passphrase string
		token      *oauth2.Token
		pass       bool
	}{
		{"pass", "good-passphrase", &oauth2.Token{AccessToken: "valid-token"}, true},
		{"bad passphrase", "bad-passphrase", &oauth2.Token{AccessToken: "valid-token"}, false},
		{"no store", "good-passphrase", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "token.enc")

			if tt.token != nil {
				writeStore := oauth2store.NewEncryptedFileTokenStore(path, "good-passphrase")
				if err := writeStore.Save(&oauth2.Token{AccessToken: "valid-token"}); err != nil {
					t.Fatalf("failed to save token: %v", err)
				}
			}

			store := oauth2store.NewEncryptedFileTokenStore(path, tt.passphrase)
			token, err := store.Load()
			if tt.pass != (err == nil) {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.pass)
			}
			if err != nil {
				return
			}
			if got := token.AccessToken; got != "valid-token" {
				t.Errorf("Token() = %v, want valid-token", got)
			}
			if store.LastSaved().IsZero() {
				t.Errorf("LastSaved() = %v, want non-zero", store.LastSaved())
			}
		})
	}
}
