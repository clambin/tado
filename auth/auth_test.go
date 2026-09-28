package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestAuthorize(t *testing.T) {
	ts := httptest.NewServer(&fakeOAuth2Server{})
	t.Cleanup(ts.Close)

	cfg := oauth2.Config{Endpoint: oauth2.Endpoint{
		DeviceAuthURL: ts.URL + "/devauth",
		TokenURL:      ts.URL + "/token",
		AuthStyle:     oauth2.AuthStyleInParams,
	}}

	token, err := Authorize(t.Context(), cfg, func(response *oauth2.DeviceAuthResponse) {})
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if !token.Valid() {
		t.Fatalf("Authorize() returned invalid token")
	}
}

func TestIsTokenRefreshable(t *testing.T) {
	tests := []struct {
		name  string
		token *oauth2.Token
		want  bool
	}{
		{
			name:  "nil",
			token: nil,
			want:  false,
		},
		{
			name:  "no refresh token",
			token: &oauth2.Token{RefreshToken: ""},
			want:  false,
		},
		{
			name: "token not refreshable",
			token: &oauth2.Token{
				RefreshToken: "refresh",
				Expiry:       time.Now().Add(-OAuth2RefreshTokenLifetime - OAuth2AccessTokenLifetime - time.Minute),
			},
			want: false,
		},
		{
			name: "token expired, but refreshable",
			token: &oauth2.Token{
				RefreshToken: "refresh",
				Expiry:       time.Now().Add(OAuth2AccessTokenLifetime - time.Minute),
			},
			want: true,
		},
		{
			name: "token valid",
			token: &oauth2.Token{
				RefreshToken: "refresh",
				Expiry:       time.Now().Add(OAuth2AccessTokenLifetime / 2),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTokenRefreshable(tt.token); got != tt.want {
				t.Errorf("IsTokenRefreshable() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeOAuth2Server struct{}

func (s *fakeOAuth2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/devauth":
		resp := oauth2.DeviceAuthResponse{
			DeviceCode:              "1234",
			UserCode:                "",
			VerificationURI:         "https://example.com",
			VerificationURIComplete: "https://example.com?code=1234",
			Expiry:                  time.Now().Add(5 * time.Minute),
			Interval:                1,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	case "/token":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(oauth2.Token{
			AccessToken:  "my-access-token",
			TokenType:    "bearer",
			RefreshToken: "my-refresh-token",
			Expiry:       time.Now().Add(OAuth2AccessTokenLifetime),
			ExpiresIn:    int64(OAuth2AccessTokenLifetime.Seconds()),
		})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
