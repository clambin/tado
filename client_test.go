package tado

import (
	"golang.org/x/oauth2"
	"net/http"
	"path/filepath"
	"testing"
)

func TestNewOAuth2Client(t *testing.T) {
	//t.Skip()

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
	resp, err := httpClient.Get("https://my.tado.com/api/v2/me")
	if err != nil {
		t.Fatalf("tado call failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}
