package oauth2store

import (
	"golang.org/x/oauth2"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestPersistentTokenSource_Token(t *testing.T) {
	validToken := &oauth2.Token{AccessToken: "valid-token", Expiry: time.Now().Add(time.Hour)}
	expiredToken := &oauth2.Token{AccessToken: "expired-token", Expiry: time.Now().Add(-time.Hour)}

	tests := []struct {
		name        string
		oauth2Token *oauth2.Token
		pass        bool
	}{
		{"no token source", nil, false},
		{"valid token", validToken, true},
		{"expired token", expiredToken, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "token.enc")
			ts := NewPersistentTokenSource(path, "secret-passphrase")

			if tt.oauth2Token != nil {
				ts.TokenSource = oauth2.StaticTokenSource(tt.oauth2Token)
			}

			got, err := ts.Token()
			if tt.pass != (err == nil) {
				t.Errorf("Token() error = %v, wantErr %v", err, tt.pass)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.oauth2Token) {
				t.Errorf("Token() = %v, want %v", got, tt.oauth2Token)
			}
			stored, err := ts.GetStoredToken(0)
			if err != nil {
				t.Errorf("Token() error = %v, wantErr %v", err, tt.pass)
			}
			if stored.AccessToken != got.AccessToken {
				t.Errorf("Token() = %v, want %v", stored, got)
			}
		})
	}
}

func TestPersistentTokenSource_GetStoredToken(t *testing.T) {
	validToken := &oauth2.Token{AccessToken: "token1", Expiry: time.Now().Add(time.Hour)}
	expiredToken := &oauth2.Token{AccessToken: "token2", Expiry: time.Now().Add(-time.Hour)}

	tests := []struct {
		name  string
		token *oauth2.Token
		age   time.Duration
		pass  bool
	}{
		{"no stored token", nil, 0, false},
		{"recent valid stored token", validToken, 0, true},
		{"recent expired stored token", expiredToken, 0, true},
		{"old stored token", validToken, 24 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "token.enc")
			ts := NewPersistentTokenSource(path, "secret-passphrase")
			if tt.token != nil {
				_ = ts.storedToken.save(tt.token)
				if tt.age != 0 {
					if err := os.Chtimes(path, time.Time{}, time.Now().Add(-tt.age)); err != nil {
						t.Fatal(err)
					}
				}
			}

			token, err := ts.GetStoredToken(time.Hour)
			if tt.pass != (err == nil) {
				t.Errorf("Token() error = %v, wantErr %v", err, tt.pass)
			}
			if err != nil {
				return
			}
			if got := token.AccessToken; got != tt.token.AccessToken {
				t.Errorf("Token() = %v, want %v", got, tt.token.AccessToken)
			}

			ts.storedToken.key = []byte("01234567890123456789012345678901")
			if _, err = ts.GetStoredToken(time.Hour); err == nil {
				t.Error("Token() error = nil, wanted not nil")
			}
		})
	}
}

func BenchmarkPersistentTokenSource(b *testing.B) {
	ts := NewPersistentTokenSource(filepath.Join(b.TempDir(), "token.enc"), "secret-passphrase")
	ts.TokenSource = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token1", Expiry: time.Now().Add(time.Hour)})
	b.ReportAllocs()
	for b.Loop() {
		_, err := ts.Token()
		if err != nil {
			b.Fatal(err)
		}
	}
}
