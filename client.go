package tado

import (
	"context"
	"golang.org/x/oauth2"
	"net/http"
)

//go:generate go tool oapi-codegen -config config.yaml https://raw.githubusercontent.com/kritsel/tado-openapispec-v2/refs/tags/v2.2025.02.03.0/tado-openapispec-v2.yaml

const ServerURL = "https://my.tado.com/api/v2"

// Config contains the oauth2 config to access the Tadoº API as per https://github.com/kritsel/tado-openapispec-v2
var Config = oauth2.Config{
	ClientID: "1bb50063-6b0c-4d11-bd99-387f4a91cc46",
	Endpoint: oauth2.Endpoint{
		DeviceAuthURL: "https://login.tado.com/oauth2/device_authorize",
		TokenURL:      "https://login.tado.com/oauth2/token",
		AuthStyle:     oauth2.AuthStyleInParams,
	},
	Scopes: []string{"offline_access"},
}

// NewOAuth2Client returns a http.Client that can access the Tadoº API.
//
// In previous versions, this call took a username & password. However, Tadoº has decided to decommission this flow (see [here]),
// in favour of the oauth2 device code authentication flow instead.
//
// Since this flow requires manual action, NewOAuth2Client returns an oauth2-enabled http.Client that works as follows:
//   - On first start-up, the client performs the device code authentication flow to get a first token.
//   - It calls deviceAuthCallback with the oauth2.DeviceAuthResponse, which contains the verification link (VerificationURIComplete).
//   - The application should display/log this link, asking the user to verify the login request.
//   - Once the client receives a token, it stores this in tokenStorePath. For security, the token is encrypted with the tokenStorePassphrase.
//   - Every time the token is renewed (10 min), the stored token is updated.
//
// When the application restarts, it reuses the stored token if the token is still valid. Otherwise, a new device code authentication flow is performed,
// and the user will need to log in again.
//
// [here]: https://github.com/wmalgadey/PyTado/issues/155
func NewOAuth2Client(ctx context.Context, tokenStorePath string, tokenStorePassphrase string, deviceAuthCallback func(response *oauth2.DeviceAuthResponse)) (*http.Client, error) {
	ts := newEncryptedTokenStore(tokenStorePath, tokenStorePassphrase)
	token, err := ts.Token()
	if err != nil {
		if token, err = deviceAuthToken(ctx, deviceAuthCallback); err == nil {
			err = ts.Store(token)
		}
	}
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(ctx, persistentTokenSource{
		TokenStore:  ts,
		TokenSource: Config.TokenSource(ctx, token),
	}), nil
}

func deviceAuthToken(ctx context.Context, deviceAuthCallback func(response *oauth2.DeviceAuthResponse)) (*oauth2.Token, error) {
	devAuth, err := Config.DeviceAuth(ctx)
	if err != nil {
		return nil, err
	}
	deviceAuthCallback(devAuth)
	return Config.DeviceAccessToken(ctx, devAuth)
}
