package tado

import (
	"golang.org/x/oauth2"
	"net/http"
	"path/filepath"
	"testing"
)

func TestNewOAuth2Client(t *testing.T) {
	t.Skip()

	httpClient, err := NewOAuth2Client(
		t.Context(),
		filepath.Join(t.TempDir(), "token.enc"),
		"my-very-secret-passphrase",
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
