package oauth2

import (
	"github.com/clambin/tado/v2"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestPersistentTokenSource(t *testing.T) {
	validToken := &oauth2.Token{AccessToken: "token1", Expiry: time.Now().Add(time.Hour)}
	expiredToken := &oauth2.Token{AccessToken: "token2", Expiry: time.Now().Add(-time.Hour)}

	tests := []struct {
		name        string
		storedToken *oauth2.Token
		oauth2Token *oauth2.Token
		pass        bool
		want        *oauth2.Token
	}{
		{"no stored token, no token source", nil, nil, false, nil},
		{"stored token, no token source", validToken, nil, true, validToken},
		{"no stored token, token source", nil, validToken, true, validToken},
		{"expired stored token, token source", expiredToken, validToken, true, validToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewEncryptedTokenStore(filepath.Join(t.TempDir(), "token.json"), "secret-passphrase")
			if tt.storedToken != nil {
				_ = s.Store(tt.storedToken)
			}

			var tokenSource oauth2.TokenSource
			if tt.oauth2Token != nil {
				tokenSource = oauth2.StaticTokenSource(tt.oauth2Token)
			}

			ts := PersistentTokenStore{TokenStore: s, TokenSource: tokenSource}
			got, err := ts.Token()
			if tt.pass != (err == nil) {
				t.Errorf("Token() error = %v, wantErr %v", err, tt.pass)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkPersistentTokenSource(b *testing.B) {
	ts := PersistentTokenStore{
		TokenStore:  NewEncryptedTokenStore(filepath.Join(b.TempDir(), "token.json"), "secret-passphrase"),
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token1", Expiry: time.Now().Add(time.Hour).UTC()}),
	}
	b.ReportAllocs()
	for b.Loop() {
		_, err := ts.Token()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestTokenStore_Token(t *testing.T) {
	validToken1 := &oauth2.Token{AccessToken: "token1", Expiry: time.Now().Add(time.Hour)}
	validToken2 := &oauth2.Token{AccessToken: "token2", Expiry: time.Now().Add(2 * time.Hour)}

	tests := []struct {
		name        string
		storedToken *oauth2.Token
		passedToken *oauth2.Token
		pass        bool
		want        *oauth2.Token
	}{
		{"no stored token, no added token", nil, nil, false, nil},
		{"valid stored token, no added token", validToken1, nil, true, validToken1},
		{"no stored token, valid added token", nil, validToken1, true, validToken1},
		{"valid stored token, valid added token", validToken2, validToken1, true, validToken1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenFile := filepath.Join(t.TempDir(), "token.json")

			store := NewEncryptedTokenStore(tokenFile, "secret-passphrase")
			if tt.storedToken != nil {
				_ = store.save(tt.storedToken)
			}

			if tt.passedToken != nil {
				_ = store.Store(tt.passedToken)
			}

			tok, err := store.Token()
			if tt.pass != (err == nil) {
				t.Errorf("expected err=%v, got err=%v", tt.pass, err)
			}
			if err != nil {
				return
			}
			if !tok.Expiry.Equal(tt.want.Expiry) || tok.AccessToken != tt.want.AccessToken {
				t.Errorf("expected tok=%v, got tok=%v", tt.want, tok)
			}
		})
	}
}

func TestCachingTokenSource_Tado(t *testing.T) {
	//t.Skip()

	// Perform the Tado device auth flow
	deviceAuthResponse, err := tado.OAuthConfig.DeviceAuth(t.Context())
	if err != nil {
		t.Fatalf("failed to get device auth response: %v", err)
	}
	t.Log("configure device auth:", deviceAuthResponse.VerificationURIComplete)

	token, err := tado.OAuthConfig.DeviceAccessToken(t.Context(), deviceAuthResponse)
	if err != nil {
		t.Fatalf("failed to get device auth token: %v", err)
	}

	persistentTokenStore := PersistentTokenStore{
		TokenStore:  NewEncryptedTokenStore(filepath.Join(os.TempDir(), "token.json"), "secret-passphrase"),
		TokenSource: tado.OAuthConfig.TokenSource(t.Context(), token),
	}
	_ = persistentTokenStore.TokenStore.Store(token)

	// Verify we can now call the Tado API
	c := oauth2.NewClient(t.Context(), persistentTokenStore)
	resp, err := c.Get("https://my.tado.com/api/v2/me")
	if err != nil {
		t.Fatalf("tado call failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}

func TestEncryptedFile(t *testing.T) {
	f := encryptedToken{
		path: filepath.Join(t.TempDir(), "token.enc"),
	}
	f.setEncryptionKey("passphrase")

	token := oauth2.Token{
		AccessToken: "token",
		Expiry:      time.Now().Add(time.Hour),
	}
	if err := f.save(&token); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	newToken, err := f.load()
	if err != nil {
		t.Fatalf("failed to load token: %v", err)
	}
	if newToken.AccessToken != token.AccessToken {
		t.Fatalf("expected token access token %s, got %s", token.AccessToken, newToken.AccessToken)
	}
}
