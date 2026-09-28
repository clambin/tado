package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

const (
	OAuth2AccessTokenLifetime  = 10 * time.Minute
	OAuth2RefreshTokenLifetime = 30 * 24 * time.Hour
)

// Authorize is a convenience function that performs the OAuth2 device authorization flow.
// The callback can be used to present the verification url to the user and ask the user to authenticate with Tadoº
// and authorize the application.
func Authorize(ctx context.Context, config oauth2.Config, callback func(response *oauth2.DeviceAuthResponse)) (*oauth2.Token, error) {
	// device auth flow
	devAuthResponse, err := config.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("DeviceAuth: %w", err)
	}
	callback(devAuthResponse)
	token, err := config.DeviceAccessToken(ctx, devAuthResponse)
	if err != nil {
		return nil, fmt.Errorf("DeviceAccessToken: %w", err)
	}
	return token, nil
}

// IsTokenRefreshable returns true if the Tadoº oauth2 token can be refreshed.
//
// Tadoº access tokens are valid for 10 minutes. Refresh tokens are valid for 30 days.
func IsTokenRefreshable(token *oauth2.Token) bool {
	return token != nil &&
		token.RefreshToken != "" &&
		token.Expiry.Add(OAuth2RefreshTokenLifetime-OAuth2AccessTokenLifetime).After(time.Now())
}
