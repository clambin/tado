package auth

import (
	"sync/atomic"
	"testing"

	"golang.org/x/oauth2"
)

func TestTokenSource_Token(t *testing.T) {
	var store fakeTokenStore
	token := &oauth2.Token{AccessToken: "token"}

	ts := TokenSource{
		TokenSource: oauth2.StaticTokenSource(token),
		TokenStore:  &store,
	}

	// get the token from the TokenSource
	got, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != token.AccessToken {
		t.Errorf("got %s, want %s", got.AccessToken, token.AccessToken)
	}
	if store.saveCount.Load() != 1 {
		t.Errorf("got %d, want 1", store.saveCount.Load())
	}

	// if the token is refreshed, the store should be updated
	token = &oauth2.Token{AccessToken: "new-token"}
	ts.TokenSource = oauth2.StaticTokenSource(token)
	got, err = ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != token.AccessToken {
		t.Errorf("got %s, want %s", got.AccessToken, token.AccessToken)
	}
	if store.saveCount.Load() != 2 {
		t.Errorf("got %d, want ", store.saveCount.Load())
	}

	// if the token remains the same, the store should not be updated
	got, err = ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != token.AccessToken {
		t.Errorf("got %s, want %s", got.AccessToken, token.AccessToken)
	}
	if store.saveCount.Load() != 2 {
		t.Errorf("got %d, want ", store.saveCount.Load())
	}
}

var _ TokenStore = (*fakeTokenStore)(nil)

type fakeTokenStore struct {
	token     atomic.Pointer[oauth2.Token]
	saveCount atomic.Int32
}

func (f *fakeTokenStore) Save(token *oauth2.Token) error {
	f.token.Store(token)
	f.saveCount.Add(1)
	return nil
}

func (f *fakeTokenStore) Load() (*oauth2.Token, error) {
	return f.token.Load(), nil
}
