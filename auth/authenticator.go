package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Authenticator implements OAuth2 for the Tado API servers
type Authenticator struct {
	HTTPClient   *http.Client
	ClientID     string
	Username     string
	Password     string
	ClientSecret string
	AuthURL      string

	lock         sync.Mutex
	accessToken  string
	refreshToken string
	expires      time.Time
}

// GetAuthToken returns a valid token to access the API server
func (a *Authenticator) GetAuthToken(ctx context.Context) (string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	var err error
	if a.refreshToken == "" {
		err = a.authenticate(ctx, "password", a.Password)
	} else if time.Now().After(a.expires) {
		if err = a.authenticate(ctx, "refresh_token", a.refreshToken); err != nil {
			// failed during refresh. reset refresh_token to force a password login
			a.refreshToken = ""
		}
	}
	return a.accessToken, err
}

// Reset forces a password-based re-authentication on the next call to GetAuthToken
func (a *Authenticator) Reset() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.refreshToken = ""
}

func (a *Authenticator) authenticate(ctx context.Context, grantType, credential string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.getURL(), strings.NewReader(a.buildForm(grantType, credential).Encode()))
	// TODO: is this needed?
	req.Header.Add("Referer", "https://my.tado.com/")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Add("Accept-Encoding", "application/json")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	var response struct {
		AccessToken  string  `json:"access_token"`
		RefreshToken string  `json:"refresh_token"`
		ExpiresIn    float64 `json:"expires_in"`
	}

	if err = json.Unmarshal(body, &response); err == nil {
		a.accessToken = response.AccessToken
		a.refreshToken = response.RefreshToken
		a.expires = time.Now().Add(time.Second * time.Duration(response.ExpiresIn))
	}
	return err
}

func (a *Authenticator) getURL() string {
	authURL := "https://auth.tado.com/"
	if a.AuthURL != "" {
		authURL = a.AuthURL
	}
	return authURL
}

func (a *Authenticator) buildForm(grantType, credential string) url.Values {
	form := make(url.Values)
	form.Add("client_id", a.ClientID)
	form.Add("client_secret", a.ClientSecret)
	form.Add("grant_type", grantType)
	form.Add(grantType, credential)
	form.Add("scope", "home.user")
	if grantType == "password" {
		form.Add("username", a.Username)
	}
	return form
}
