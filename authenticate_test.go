package tado

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestAuthenticator_AuthHeaders(t *testing.T) {
	server := authServer{}
	testServer := httptest.NewServer(http.HandlerFunc(server.authHandler))

	auth := authenticator{
		HTTPClient: &http.Client{},
		Username:   "user@examle.com",
		Password:   "some-password",
		AuthURL:    testServer.URL,
	}

	headers, err := auth.AuthHeaders(context.Background())
	require.NoError(t, err)
	token := headers.Get("Authorization")
	assert.NotZero(t, token)

	// if already authenticated, we should re-use the token
	headers, err = auth.AuthHeaders(context.Background())
	require.NoError(t, err)
	newToken := headers.Get("Authorization")
	assert.NotZero(t, newToken)
	assert.Equal(t, token, newToken)

	// expire token on client side. we should get a new token.
	auth.Expires = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	headers, err = auth.AuthHeaders(context.Background())
	require.NoError(t, err)
	newToken = headers.Get("Authorization")
	assert.NotZero(t, newToken)
	assert.NotEqual(t, token, newToken)

	// auth server is failing
	server.fail = true
	auth.Expires = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = auth.AuthHeaders(context.Background())
	require.Error(t, err)

	// auth server is down
	testServer.Close()
	auth.Expires = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = auth.AuthHeaders(context.Background())
	assert.Error(t, err)
}

func TestAuthenticator_Reset(t *testing.T) {
	server := authServer{}
	testServer := httptest.NewServer(http.HandlerFunc(server.authHandler))

	auth := authenticator{
		HTTPClient: &http.Client{},
		Username:   "user@examle.com",
		Password:   "some-password",
		AuthURL:    testServer.URL,
	}

	headers, err := auth.AuthHeaders(context.Background())
	require.NoError(t, err)
	token := headers.Get("Authorization")
	assert.NotZero(t, token)

	auth.Reset()

	assert.Zero(t, auth.refreshToken)
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
