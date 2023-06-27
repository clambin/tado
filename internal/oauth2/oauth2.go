package oauth2

import (
	"context"
	"golang.org/x/oauth2"
	"net/http"
)

func NewClient(ctx context.Context, username, password, clientSecret string) (*http.Client, error) {
	if clientSecret == "" {
		clientSecret = "wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc"
	}
	cfg := oauth2.Config{
		ClientID:     "tado-web-app",
		ClientSecret: clientSecret,
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
