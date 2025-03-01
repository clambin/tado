package tado

import (
	"golang.org/x/oauth2"
)

//go:generate go tool oapi-codegen -config config.yaml https://raw.githubusercontent.com/kritsel/tado-openapispec-v2/refs/tags/v2.2025.02.03.0/tado-openapispec-v2.yaml

const ServerURL = "https://my.tado.com/api/v2"

// OAuthConfig as per https://github.com/kritsel/tado-openapispec-v2
var OAuthConfig = oauth2.Config{
	ClientID: "1bb50063-6b0c-4d11-bd99-387f4a91cc46",
	Endpoint: oauth2.Endpoint{
		DeviceAuthURL: "https://login.tado.com/oauth2/device_authorize",
		TokenURL:      "https://login.tado.com/oauth2/token",
		AuthStyle:     oauth2.AuthStyleInParams,
	},
	Scopes: []string{"offline_access"},
}

/*
func NewOAuth2Client(ctx context.Context, username string, password string) (*http.Client, error) {
	tok, err := OAuthConfig.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(ctx, OAuthConfig.TokenSource(ctx, tok)), nil
}

func MustNewOAuth2Client(ctx context.Context, username string, password string) *http.Client {
	c, err := NewOAuth2Client(ctx, username, password)
	if err != nil {
		panic(err)
	}
	return c
}
*/
