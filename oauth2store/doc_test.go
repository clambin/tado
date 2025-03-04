package oauth2store_test

import (
	"context"
	"fmt"
	"github.com/clambin/tado/v2/oauth2store"
	"golang.org/x/oauth2"
	"time"
)

func ExampleTokenSource() {
	var cfg oauth2.Config // application-specific oauth2 configuration
	ctx := context.Background()

	// create a store to save the token
	store := oauth2store.NewEncryptedFileTokenStore("token.enc", "my-very-secret-passphrase", 24*time.Hour)
	token, err := store.Load()
	if err != nil {
		// store does not contain a valid token. let's create one ...
		var devAuthResponse *oauth2.DeviceAuthResponse
		if devAuthResponse, err = cfg.DeviceAuth(ctx); err != nil {
			panic(err)
		}
		fmt.Println("Confirm login: ", devAuthResponse.VerificationURIComplete)
		if token, err = cfg.DeviceAccessToken(ctx, devAuthResponse); err != nil {
			panic(err)
		}
	}

	// now that we have a token, create a TokenSource that saves the token to our store whenever it changes:
	pts := oauth2store.TokenSource{
		TokenSource: cfg.TokenSource(ctx, token),
		TokenStore:  store,
	}
	_ = oauth2.NewClient(ctx, &pts)
}
