package tado

import (
	"context"
	"golang.org/x/oauth2"
	"net/http"
)

////go:generate oapi-codegen -config config.yaml https://raw.githubusercontent.com/kritsel/tado-openapispec-v2/refs/tags/v2.2024.09.22.0/tado-openapispec-v2.yaml

////go:generate oapi-codegen -config config.yaml ../tado-openapispec-v2/tado-openapispec-v2.yaml

//go:generate oapi-codegen -config config.yaml https://raw.githubusercontent.com/clambin/tado-openapispec-v2/refs/heads/fix-weather/tado-openapispec-v2.yaml
const ServerURL = "https://my.tado.com/api/v2"

func NewOAuth2Client(ctx context.Context, username string, password string) (*http.Client, error) {
	cfg := oauth2.Config{
		ClientID:     "public-api-preview",
		ClientSecret: "4HJGRffVR8xb3XdEUQpjgZ1VplJi6Xgw",
		Endpoint: oauth2.Endpoint{
			//AuthURL:   "https://auth.tado.com/oauth/auth",
			TokenURL:  "https://auth.tado.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		Scopes: []string{"home.user"},
	}
	tok, err := cfg.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(ctx, cfg.TokenSource(ctx, tok)), nil
}

func MustNewOAuth2Client(ctx context.Context, username string, password string) *http.Client {
	c, err := NewOAuth2Client(ctx, username, password)
	if err != nil {
		panic(err)
	}
	return c
}
