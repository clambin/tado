package tado

import (
	"context"
	"golang.org/x/oauth2"
	"net/http"
)

////go:generate oapi-codegen -config config.yaml ../tado-openapispec-v2/tado-openapispec-v2.yaml
//go:generate oapi-codegen -config config.yaml https://raw.githubusercontent.com/kritsel/tado-openapispec-v2/refs/tags/v2.2024.12.30.1/tado-openapispec-v2.yaml

const ServerURL = "https://my.tado.com/api/v2"

// tadoOAuthConfig as per https://github.com/kritsel/tado-openapispec-v2
var tadoOAuthConfig = oauth2.Config{
	ClientID:     "public-api-preview",
	ClientSecret: "4HJGRffVR8xb3XdEUQpjgZ1VplJi6Xgw",
	Endpoint: oauth2.Endpoint{
		//AuthURL:   "https://auth.tado.com/oauth/auth",
		TokenURL:  "https://auth.tado.com/oauth/token",
		AuthStyle: oauth2.AuthStyleInHeader,
	},
	Scopes: []string{"home.user"},
}

func NewOAuth2Client(ctx context.Context, username string, password string) (*http.Client, error) {
	tok, err := tadoOAuthConfig.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(ctx, tadoOAuthConfig.TokenSource(ctx, tok)), nil
}

func MustNewOAuth2Client(ctx context.Context, username string, password string) *http.Client {
	c, err := NewOAuth2Client(ctx, username, password)
	if err != nil {
		panic(err)
	}
	return c
}
