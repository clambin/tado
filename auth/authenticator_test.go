package auth

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestAuthenticator_GetAuthToken(t *testing.T) {
	server := authServer{}
	testServer := httptest.NewServer(http.HandlerFunc(server.authHandler))

	auth := Authenticator{
		HTTPClient: &http.Client{},
		Username:   "user@examle.com",
		Password:   "some-password",
		AuthURL:    testServer.URL + "/oauth/token",
	}

	token, err := auth.GetAuthToken(context.Background())
	require.NoError(t, err)
	assert.NotZero(t, token)

	// if already authenticated, we should re-use the token
	newToken, err := auth.GetAuthToken(context.Background())
	require.NoError(t, err)
	assert.NotZero(t, newToken)
	assert.Equal(t, token, newToken)

	// expire token on client side. we should get a new token.
	auth.expires = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	newToken, err = auth.GetAuthToken(context.Background())
	require.NoError(t, err)
	assert.NotZero(t, newToken)
	assert.NotEqual(t, token, newToken)

	// reset forces a new password-based login
	auth.Reset()
	assert.Zero(t, auth.refreshToken)
	newToken, err = auth.GetAuthToken(context.Background())
	require.NoError(t, err)
	assert.NotZero(t, newToken)
	assert.Equal(t, token, newToken)

	// auth server is failing
	server.fail = true
	auth.expires = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = auth.GetAuthToken(context.Background())
	require.Error(t, err)

	// auth server is down
	testServer.Close()
	_, err = auth.GetAuthToken(context.Background())
	assert.Error(t, err)
}

func TestAuthenticator_buildForm(t *testing.T) {
	tests := []struct {
		name       string
		grantType  string
		credential string
		want       string
	}{
		{
			name:       "password",
			grantType:  "password",
			credential: "my-password",
			want:       "client_id=foo&client_secret=123&grant_type=password&password=my-password&scope=home.user&username=bar%40example.com",
		},
		{
			name:       "refresh_token",
			grantType:  "refresh_token",
			credential: "my-token",
			want:       "client_id=foo&client_secret=123&grant_type=refresh_token&refresh_token=my-token&scope=home.user",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := Authenticator{ClientID: "foo", ClientSecret: "123", Username: "bar@example.com", Password: "my-password2"}
			assert.Equal(t, tt.want, auth.buildForm(tt.grantType, tt.credential).Encode())
		})
	}
}

// authServer implements an authentication server
type authServer struct {
	counter      int
	accessToken  string
	refreshToken string
	expires      time.Time
	fail         bool
	failRefresh  bool
}

func (server *authServer) authHandler(w http.ResponseWriter, req *http.Request) {
	if server.fail {
		http.Error(w, "server is having issues", http.StatusInternalServerError)
		return
	}

	if req.URL.Path != "/oauth/token" {
		http.Error(w, "endpoint not implemented", http.StatusNotFound)
		return
	}

	response, ok := server.handleAuthentication(req)

	if ok == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
	} else {
		_, _ = w.Write([]byte(response))
	}
}

func (server *authServer) handleAuthentication(req *http.Request) (response string, ok bool) {
	const authResponse = `{
  		"access_token":"%s",
  		"token_type":"bearer",
  		"refresh_token":"%s",
  		"expires_in":%d,
  		"scope":"home.user",
  		"jti":"jti"
	}`

	grantType := getGrantType(req.Body)

	if grantType == "refresh_token" {
		if server.failRefresh {
			return "test server in failRefresh mode", false
		}
		server.counter++
	} else {
		server.counter = 1
	}

	server.accessToken = fmt.Sprintf("token_%d", server.counter)
	server.refreshToken = server.accessToken
	server.expires = time.Now().Add(20 * time.Second)

	return fmt.Sprintf(authResponse, server.accessToken, server.refreshToken, 20), true
}

func getGrantType(body io.Reader) string {
	content, _ := io.ReadAll(body)
	if params, err := url.ParseQuery(string(content)); err == nil {
		if tokenType, ok := params["grant_type"]; ok == true {
			return tokenType[0]
		}
	}
	panic("grant_type not found in body")
}

func TestAuthenticator_GetAuthToken_E2E(t *testing.T) {
	username := os.Getenv("TADO_USERNAME")
	password := os.Getenv("TADO_PASSWORD")

	if username == "" || password == "" {
		t.Skip("environment not set. skipping ...")
	}

	a := Authenticator{
		HTTPClient:   http.DefaultClient,
		ClientID:     "foo",
		ClientSecret: "wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc",
		Username:     username,
		Password:     password,
		AuthURL:      "https://auth.tado.com/oauth/token",
	}

	ctx := context.Background()
	token, err := a.GetAuthToken(ctx)
	require.NoError(t, err)
	t.Log(token)
}
