package oauth2store_test

import (
	"sync/atomic"
	"testing"

	"github.com/clambin/tado/v2/oauth2store"
	"golang.org/x/oauth2"
)

func TestTokenSource_Token(t *testing.T) {
	var store fakeTokenStore
	token := &oauth2.Token{AccessToken: "token"}

	ts := oauth2store.TokenSource{
		TokenSource: oauth2.StaticTokenSource(token),
		TokenStore: oauth2store.TokenStore{
			Storer: &store,
		},
	}

	// get token twice. OnChangedToken should only be called the first time.
	for range 2 {
		got, err := ts.Token()
		if err != nil {
			t.Fatal(err)
		}
		if got.AccessToken != token.AccessToken {
			t.Errorf("got %s, want %s", got.AccessToken, token.AccessToken)
		}
	}

	// store now holds token
	loaded, err := ts.TokenStore.Load()
	if err != nil {
		t.Fatalf("failed to load token: %s", err)
	}
	if loaded.AccessToken != token.AccessToken {
		t.Errorf("got %s, want %s", loaded.AccessToken, token.AccessToken)
	}
	if store.saveCount.Load() != 1 {
		t.Errorf("got %d, want 1", store.saveCount.Load())
	}

	// new token. OnChangedToken should be called again.
	token = &oauth2.Token{AccessToken: "new-token"}
	ts.TokenSource = oauth2.StaticTokenSource(token)
	got, err := ts.Token()
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

var _ oauth2store.Storer = (*fakeTokenStore)(nil)

type fakeTokenStore struct {
	token     atomic.Value
	saveCount atomic.Int32
}

func (f *fakeTokenStore) Save(bytes []byte) error {
	f.token.Store(bytes)
	f.saveCount.Add(1)
	return nil
}

func (f *fakeTokenStore) Load() ([]byte, error) {
	return f.token.Load().([]byte), nil
}
