// Package tado provides an API Client for the Tadoº smart thermostat devices
//
// Using this package typically involves creating an APIClient as follows:
//
//	client := tado.New("your-tado-username", "your-tado-password", "your-tado-secret")
//
// Once a client has been created, you can query tado.com for information about your different Tado devices.
// Currently, the following endpoints are supported:
//
//	GetZones:                   get the different zones (rooms) defined in your home
//	GetZoneInfo:                get metrics for a specified zone in your home
//	GetWeatherInfo:             get overall weather information
//	GetMobileDevices:           get status of each registered mobile device
//	SetZoneOverlay              set a permanent overlay for a zone
//	SetZoneOverlayWithDuration  set a temporary overlay for a zone
//	DeleteZoneOverlay           delete the overlay for a zone
package tado

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/clambin/tado/auth"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Common Tado data structures

// Temperature contains a temperature in degrees Celsius
type Temperature struct {
	Celsius float64 `json:"celsius"`
}

// Percentage contains a percentage (0-100%)
type Percentage struct {
	Percentage float64 `json:"percentage"`
}

// Value contains a string value
type Value struct {
	Value string `json:"value"`
}

// API for the Tado APIClient.
//
//go:generate mockery --name API
type API interface {
	GetZones(context.Context) ([]Zone, error)
	GetZoneInfo(context.Context, int) (ZoneInfo, error)
	GetWeatherInfo(context.Context) (WeatherInfo, error)
	GetMobileDevices(context.Context) ([]MobileDevice, error)
	SetZoneOverlay(context.Context, int, float64) error
	SetZoneOverlayWithDuration(context.Context, int, float64, time.Duration) error
	DeleteZoneOverlay(context.Context, int) error
}

// APIClient represents a Tado API client.
type APIClient struct {
	// authenticator handles logging in to the Tado server
	authenticator
	// HTTPClient is used to perform HTTP requests
	HTTPClient *http.Client

	lock   sync.RWMutex
	apiURL string
	HomeID int
}

type authenticator interface {
	GetAuthToken(ctx context.Context) (token string, err error)
	Reset()
}

// New creates a new client
//
// clientSecret can typically be left blank.  If the default secret does not work, your client secret can be found by visiting https://my.tado.com/webapp/env.js after logging in to https://my.tado.com
func New(username, password, clientSecret string) *APIClient {
	if clientSecret == "" {
		clientSecret = "wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc"
	}

	return &APIClient{
		authenticator: &auth.Authenticator{
			HTTPClient:   http.DefaultClient,
			ClientID:     "tado-web-app",
			ClientSecret: clientSecret,
			Username:     username,
			Password:     password,
			AuthURL:      "https://auth.tado.com/oauth/token",
		},
		HTTPClient: http.DefaultClient,
		apiURL:     "https://my.tado.com",
	}
}

// apiV2URL returns a API v2 URL
func (client *APIClient) apiV2URL(endpoint string) string {
	return client.apiURL + "/api/v2/homes/" + strconv.Itoa(client.HomeID) + endpoint
}

// getHomeID gets the user's Home ID
//
// Called by Initialize, so doesn't need to be called by the calling application.
func (client *APIClient) getHomeID(ctx context.Context) (err error) {
	client.lock.Lock()
	homeID := client.HomeID
	client.lock.Unlock()

	if homeID > 0 {
		return nil
	}

	var meResponse struct {
		HomeID int `json:"homeId"`
	}

	if err = client.call(ctx, http.MethodGet, client.apiURL+"/api/v1/me", bytes.NewBufferString(""), &meResponse); err != nil {
		return
	}

	client.lock.Lock()
	client.HomeID = meResponse.HomeID
	client.lock.Unlock()
	return
}

// Initialize sets up the client to call the various APIs, i.e. authenticates with tado.com,
// retrieving/updating the Access Token required for the API functions, and retrieving the
// user's Home ID.
//
// Each API function calls this before invoking the API, so normally this doesn't need to be
// called by the calling application.
func (client *APIClient) initialize(ctx context.Context) (err error) {
	return client.getHomeID(ctx)
}

func (client *APIClient) call(ctx context.Context, method string, url string, payload io.Reader, response any) error {
	req, err := client.buildRequest(ctx, method, url, payload)
	if err != nil {
		return err
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength != 0 && response != nil {
			err = json.Unmarshal(body, response)
		}
	case http.StatusNoContent:
	case http.StatusForbidden, http.StatusUnauthorized:
		// we're authenticated, but still got forbidden.
		// force password login to get a new token.
		client.authenticator.Reset()
		err = errors.New(resp.Status)
	default:
		err = errors.New(resp.Status)
	}

	return err
}

func (client *APIClient) buildRequest(ctx context.Context, method string, path string, payload io.Reader) (*http.Request, error) {
	token, err := client.authenticator.GetAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}
	req, _ := http.NewRequestWithContext(ctx, method, path, payload)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+token)

	return req, nil
}
