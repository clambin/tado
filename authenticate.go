package tado

import (
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Authenticator provides authentication services for tado.com
//go:generate mockery --name Authenticator
type Authenticator interface {
	AuthHeaders(ctx context.Context) (header http.Header, err error)
	Reset()
}

// authenticator handles authentication for the Tado API. Tado uses a two-phased approach to authentication:
//
// The first authentication is based on the user password.  If successful, it returns an accessToken, for use in subsequent API calls, and a refreshToken.
//
// The accessToken has a lifetime, as indicated by the Expires field. When the accessToken expires, it must be renewed by authenticating again with the refreshToken.
//
// Therefore, as long as accessToken gets renewed before it expires, password-based authentication is only needed on the first call.
type authenticator struct {
	// HTTPClient for use in accessing the authentication API
	HTTPClient *http.Client
	// Username for your Tado account
	Username string
	// Password for your Tado account
	Password string
	// ClientSecret is used during authentication. Can normally be left blank. Authenticator will use the default secret in that case.
	// If this does not work, log into tado.com and visit [https://my.tado.com/webapp/env.js](https://my.tado.com/webapp/env.js).
	// The client secret can be found in clientSecret key in the oauth section:
	ClientSecret string
	// AuthURL can be left blank.  Only exported for unit test purposes
	AuthURL string

	lock         sync.RWMutex
	accessToken  string
	refreshToken string
	// Expires does not need to be provided. Only exported for unit test purposes
	Expires time.Time
}

// AuthHeaders authenticates with the Tado servers and returns the required authentication headers for an API call
func (auth *authenticator) AuthHeaders(ctx context.Context) (header http.Header, err error) {
	auth.lock.Lock()
	defer auth.lock.Unlock()

	err = auth.authenticate(ctx)
	if err != nil {
		return
	}

	header = http.Header{}
	header.Add("Authorization", "Bearer "+auth.accessToken)
	return
}

// Reset forces a password-based re-authentication on the next call to AuthHeaders
func (auth *authenticator) Reset() {
	auth.lock.Lock()
	defer auth.lock.Unlock()
	auth.refreshToken = ""
}

func (auth *authenticator) authenticate(ctx context.Context) (err error) {
	if auth.ClientSecret == "" {
		auth.ClientSecret = "wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc"
	}

	if auth.refreshToken == "" {
		return auth.doAuthentication(ctx, "password", auth.Password)
	}

	if time.Now().After(auth.Expires) {
		return auth.doAuthentication(ctx, "refresh_token", auth.refreshToken)
	}

	return
}

func (auth *authenticator) doAuthentication(ctx context.Context, grantType, credential string) (err error) {
	log.WithField("grant_type", grantType).Debug("authenticating")

	form := url.Values{}
	form.Add("client_id", "tado-web-app")
	form.Add("client_secret", auth.ClientSecret)
	form.Add("grant_type", grantType)
	form.Add(grantType, credential)
	form.Add("scope", "home.user")
	if grantType == "password" {
		form.Add("username", auth.Username)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, auth.authURL()+"/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Add("Referer", "https://my.tado.com/")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	var resp *http.Response
	resp, err = auth.HTTPClient.Do(req)

	if err == nil {
		if resp.StatusCode == http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)

			var response interface{}
			if err = json.Unmarshal(body, &response); err == nil {
				m := response.(map[string]interface{})
				auth.accessToken = m["access_token"].(string)
				auth.refreshToken = m["refresh_token"].(string)
				auth.Expires = time.Now().Add(time.Second * time.Duration(m["expires_in"].(float64)))
			}
		} else {
			err = errors.New(resp.Status)
		}
		_ = req.Body.Close()
	}

	if err != nil && grantType == "refresh_token" {
		// failed during refresh. reset refresh_token to force a password login
		auth.refreshToken = ""
	}
	log.WithError(err).WithField("expires", auth.Expires).Debug("authenticated")

	return err
}

// authURL returns the URL of the authentication server
func (auth *authenticator) authURL() (authURL string) {
	authURL = baseAuthURL
	if auth.AuthURL != "" {
		authURL = auth.AuthURL

	}
	return
}
