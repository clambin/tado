package tado

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/clambin/tado/v2/auth"
	"golang.org/x/oauth2"
)

func TestNewTadoHTTPClient(t *testing.T) {
	t.Skip() // interactive test

	store := auth.NewEncryptedFileTokenStore(filepath.Join(t.TempDir(), "token.enc"), "my-very-secret-passphrase")

	httpClient, err := NewTadoHTTPClient(
		t.Context(),
		store,
		func(response *oauth2.DeviceAuthResponse) {
			t.Logf("confirm login request: %+v", response.VerificationURIComplete)
		},
	)
	if err != nil {
		t.Fatalf("NewOAuth2Client failed: %v", err)
	}

	client, err := NewClientWithResponses(ServerURL, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewOAuth2Client failed: %v", err)
	}

	resp, err := client.GetMeWithResponse(t.Context())
	if err != nil {
		t.Fatalf("GetMe() failed: %v", err)
	}
	if code := resp.StatusCode(); code != http.StatusOK {
		t.Fatalf("GetMe() failed: %v", code)
	}
}
