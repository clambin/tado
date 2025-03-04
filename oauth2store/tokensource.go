package oauth2store

import (
	"golang.org/x/oauth2"
	"sync/atomic"
)

var _ oauth2.TokenSource = &TokenSource{}

// TokenSource is an [oauth2.TokenSource] that saves the token whenever the underlying TokenSource returns a new token.
//
// The main use case is to persist tokens that require manual action to create (e.g. device authentication flow), so that
// a previously created token may be reused if the application is restarted.
//
// [oauth2.TokenSource]: https://pkg.go.dev/golang.org/x/oauth2#TokenSource
type TokenSource struct {
	oauth2.TokenSource
	TokenStore
	currentToken atomic.Value
}

// Token implements interface oauth2.TokenSource. It gets a token from the underlying TokenSource and,
// if the token is new, stores the token in its TokenStore.
func (ts *TokenSource) Token() (token *oauth2.Token, err error) {
	// get an active token from the token source
	if token, err = ts.TokenSource.Token(); err != nil {
		return nil, err
	}

	// save it if needed
	currentToken := ts.currentToken.Load()
	if currentToken == nil || currentToken.(*oauth2.Token).AccessToken != token.AccessToken {
		if err = ts.TokenStore.Save(token); err == nil {
			ts.currentToken.Store(token)
		}
	}
	return token, err
}
