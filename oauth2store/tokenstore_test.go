package oauth2store_test

import (
	"github.com/clambin/tado/v2/oauth2store"
	"golang.org/x/oauth2"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEncryptedFileStore_Load(t *testing.T) {
	type storeConfig struct {
		passphrase string
		maxAge     time.Duration
	}
	tests := []struct {
		name        string
		writeConfig *storeConfig
		readConfig  storeConfig
		pass        bool
	}{
		{"pass", &storeConfig{"good-passphrase", 0}, storeConfig{"good-passphrase", time.Hour}, true},
		{"bad passphrase", &storeConfig{"good-passphrase", 0}, storeConfig{"bad-passphrase", time.Hour}, false},
		{"no store", nil, storeConfig{"good-passphrase", time.Hour}, false},
		{"old store", &storeConfig{"good-passphrase", 2 * time.Hour}, storeConfig{"good-passphrase", time.Hour}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "token.enc")

			if tt.writeConfig != nil {
				writeStore := oauth2store.NewEncryptedFileTokenStore(path, tt.writeConfig.passphrase, tt.writeConfig.maxAge)
				if err := writeStore.Save(&oauth2.Token{AccessToken: "valid-token"}); err != nil {
					t.Fatalf("failed to save token: %v", err)
				}
				err := os.Chtimes(path, time.Time{}, time.Now().Add(-tt.writeConfig.maxAge))
				if err != nil {
					t.Fatalf("failed to set token file age: %v", err)
				}
			}

			store := oauth2store.NewEncryptedFileTokenStore(path, tt.readConfig.passphrase, tt.readConfig.maxAge)
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
		})
	}
}
