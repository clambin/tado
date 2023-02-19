package tado

import (
	"fmt"
	"net/http"
)

var _ http.RoundTripper = &roundTripper{}

type roundTripper struct {
	authenticator
}

func (r roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	token, err := r.authenticator.GetAuthToken(request.Context())
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)

	return http.DefaultTransport.RoundTrip(request)
}
