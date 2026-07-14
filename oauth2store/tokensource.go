package oauth2store

import (
	"sync/atomic"

	"golang.org/x/oauth2"
)

var _ oauth2.TokenSource = &TokenSource{}

// TokenSource is an [oauth2.TokenSource] that saves the token whenever the underlying oauth2.TokenSource returns a new token.
//
// The main use case is to persist tokens that require manual action to create (e.g., device authentication flow),
// so that a previously created token may be reused if the application is restarted.
//
// [oauth2.TokenSource]: https://pkg.go.dev/golang.org/x/oauth2#TokenSource
type TokenSource struct {
	oauth2.TokenSource
	TokenStore
	currentToken atomic.Pointer[oauth2.Token]
}

// Token implements the oauth2.TokenSource interface. It gets a token from the underlying TokenSource and,
// if the token is new, stores the token in its TokenStore.
func (ts *TokenSource) Token() (token *oauth2.Token, err error) {
	// get an active token from the underlying token source
	if token, err = ts.TokenSource.Token(); err != nil {
		return nil, err
	}

	// if the token is new or has changed, save it so we can load it on startup
	currentToken := ts.currentToken.Load()
	if currentToken == nil || currentToken.AccessToken != token.AccessToken {
		if err = ts.Save(token); err == nil {
			ts.currentToken.Store(token)
		}
	}
	return token, err
}
