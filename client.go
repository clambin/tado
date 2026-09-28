package tado

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/clambin/tado/v2/auth"
	"github.com/clambin/tado/v2/oauth2store"
	"golang.org/x/oauth2"
)

//go:generate go tool oapi-codegen -config config.yaml https://raw.githubusercontent.com/kritsel/tado-openapispec-v2/refs/tags/v2.2025.02.03.0/tado-openapispec-v2.yaml

const (
	ServerURL = "https://my.tado.com/api/v2"
)

// Config contains the oauth2 config to access the Tadoº API, as per https://github.com/wmalgadey/PyTado/issues/155
var Config = oauth2.Config{
	ClientID: "1bb50063-6b0c-4d11-bd99-387f4a91cc46",
	Endpoint: oauth2.Endpoint{
		DeviceAuthURL: "https://login.tado.com/oauth2/device_authorize",
		TokenURL:      "https://login.tado.com/oauth2/token",
		AuthStyle:     oauth2.AuthStyleInParams,
	},
	Scopes: []string{"offline_access"},
}

// NewOAuth2Client returns an http.Client that can access the Tadoº API.
//
// Deprecated: use [NewTadoHTTPClient] instead.
func NewOAuth2Client(ctx context.Context, tokenStorePath string, tokenStorePassphrase string, deviceAuthCallback func(response *oauth2.DeviceAuthResponse)) (client *http.Client, err error) {
	// store to save our token
	store := oauth2store.NewEncryptedFileTokenStore(tokenStorePath, tokenStorePassphrase)
	var token *oauth2.Token
	// Tado refresh token is valid for 30 days
	if time.Since(store.LastSaved()) < auth.OAuth2RefreshTokenLifetime {
		token, _ = store.Load()
	}
	// if our store doesn't have a valid token, ask the user to log in
	if token == nil {
		var devAuthResponse *oauth2.DeviceAuthResponse
		if devAuthResponse, err = Config.DeviceAuth(ctx); err != nil {
			return nil, fmt.Errorf("DevAuth: %w", err)
		}
		deviceAuthCallback(devAuthResponse)
		if token, err = Config.DeviceAccessToken(ctx, devAuthResponse); err != nil {
			return nil, fmt.Errorf("DeviceAccessToken: %w", err)
		}
	}
	pts := oauth2store.TokenSource{
		TokenSource: Config.TokenSource(ctx, token),
		TokenStore:  store,
	}
	return oauth2.NewClient(ctx, &pts), nil
}

// NewTadoHTTPClient returns an http.Client that can access the Tadoº API.
//
// In previous versions, this call took a username and password. However, Tadoº has decided to decommission this flow (see [here]),
// in favor of the oauth2 device code authentication flow instead.
//
// Since this flow requires manual action, NewTadoHTTPClient returns an oauth2-enabled http.Client that works as follows:
//   - On first start-up, the client performs the device code authentication flow to get a first token.
//   - It calls deviceAuthCallback with the oauth2.DeviceAuthResponse, which contains the verification link (VerificationURIComplete).
//   - The application should display/log this link, asking the user to verify the login request.
//   - Once the client receives a token, it stores this in an auth.TokenStore. For security, the token is encrypted with the tokenStorePassphrase.
//   - Every time the token is renewed (10 min), the stored token is written to disk.
//
// When the application restarts, it reuses the stored token if the token is still valid. Otherwise, a new device code authentication flow is performed,
// and the user will need to log in again.
//
// [here]: https://github.com/wmalgadey/PyTado/issues/155
func NewTadoHTTPClient(ctx context.Context, tokenStore auth.TokenStore, deviceAuthCallback func(response *oauth2.DeviceAuthResponse)) (*http.Client, error) {
	// check the token in the token store
	token, err := tokenStore.Load()
	// Tado tokens get refreshed every 10 minutes. The refresh token is valid for 30 days.
	if errors.Is(err, auth.ErrNoTokenFound) || !auth.IsTokenRefreshable(token) {
		// no refreshable token found. start the device auth flow
		token, err = auth.Authorize(ctx, Config, deviceAuthCallback)
		if err != nil {
			return nil, fmt.Errorf("DeviceAccessToken: %w", err)
		}
	}
	// create an oauth2 TokenSource that refreshes the token with Tadoº and stores the token whenever it is refreshed
	ts := &auth.TokenSource{
		TokenSource: Config.TokenSource(ctx, token),
		TokenStore:  tokenStore,
	}
	return oauth2.NewClient(ctx, ts), nil
}
