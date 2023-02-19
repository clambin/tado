package tado

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoundTripper_RoundTrip(t *testing.T) {
	s := httptest.NewServer(authenticationHandler("1234")(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})))
	defer s.Close()

	auth := fakeAuthenticator{Token: "1234"}
	c := http.Client{Transport: roundTripper{authenticator: &auth}}
	req, _ := http.NewRequest(http.MethodGet, s.URL, nil)
	resp, err := c.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, auth.set)

	auth = fakeAuthenticator{Token: "4321"}
	c = http.Client{Transport: roundTripper{authenticator: &auth}}
	req, _ = http.NewRequest(http.MethodGet, s.URL, nil)
	resp, err = c.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.False(t, auth.set)
}
