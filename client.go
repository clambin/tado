// Package tado provides an API Client for the Tadoº smart thermostat devices
//
// Using this package typically involves creating an APIClient as follows:
//
// 		client := tado.APIClient{
//			Authenticator: tado.Authenticator{
//				HTTPClient: &http.Client{},
//				Username:   "your-tado-username",
//				Password:   "your-tado-password",
// 			},
//			HTTPClient: &http.Client{},
//     }
//
// Once a client has been created, you can query tado.com for information about your different Tado devices.
// Currently, the following endpoints are supported:
//
//   GetZones:                   get the different zones (rooms) defined in your home
//   GetZoneInfo:                get metrics for a specified zone in your home
//   GetWeatherInfo:             get overall weather information
//   GetMobileDevices:           get status of each registered mobile device
//   SetZoneOverlay              set a permanent overlay for a zone
//   SetZoneOverlayWithDuration  set a temporary overlay for a zone
//   DeleteZoneOverlay           delete the overlay for a zone
//
package tado

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
//
// 		client := tado.APIClient{
//			Authenticator: tado.Authenticator{
//				HTTPClient: &http.Client{},
//				Username:   "your-tado-username",
//				Password:   "your-tado-password",
// 			},
//			HTTPClient: &http.Client{},
//     }
//
// If the default Client Secret does not work, you can provide your own secret:
// 		client := tado.APIClient{
//			Authenticator: tado.Authenticator{
//				HTTPClient: &http.Client{},
//				Username:   "your-tado-username",
//				Password:   "your-tado-password",
//				ClientSecret: "your-client-secret",
// 			},
//			HTTPClient: &http.Client{},
//     }
//
// where your-client-secret can be found by visiting https://my.tado.com/webapp/env.js after logging in to https://my.tado.com
//
type APIClient struct {
	// Authenticator handles logging in to the Tado server
	Authenticator Authenticator
	// HTTPClient is used to perform HTTP requests
	HTTPClient *http.Client
	// APIURL can be left blank. Only exposed for unit tests.
	APIURL string

	HomeID int
	lock   sync.RWMutex
}

// New creates a new client.
func New(username, password, clientSecret string) *APIClient {
	return &APIClient{
		Authenticator: &authenticator{
			HTTPClient:   &http.Client{},
			Username:     username,
			Password:     password,
			ClientSecret: clientSecret,
			AuthURL:      baseAuthURL,
		},
		HTTPClient: &http.Client{},
		APIURL:     baseAPIURL,
	}
}

const baseAPIURL = "https://my.tado.com"
const baseAuthURL = "https://auth.tado.com"

// apiV2URL returns a API v2 URL
func (client *APIClient) apiV2URL(endpoint string) string {
	return client.APIURL + "/api/v2/homes/" + strconv.Itoa(client.HomeID) + endpoint
}

// getHomeID gets the user's Home ID
//
// Called by Initialize, so doesn't need to be called by the calling application.
func (client *APIClient) getHomeID(ctx context.Context) error {
	client.lock.Lock()
	homeID := client.HomeID
	client.lock.Unlock()

	if homeID > 0 {
		return nil
	}

	body, err := client.call(ctx, http.MethodGet, client.APIURL+"/api/v1/me", "")

	if err == nil {
		var resp interface{}
		if err = json.Unmarshal(body, &resp); err == nil {
			m := resp.(map[string]interface{})
			client.lock.Lock()
			client.HomeID = int(m["homeId"].(float64))
			client.lock.Unlock()
		}
	}
	return err
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

func (client *APIClient) call(ctx context.Context, method string, apiURL string, payload string) (response []byte, err error) {
	var req *http.Request
	req, err = client.buildRequest(ctx, method, apiURL, payload)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = client.HTTPClient.Do(req)
	if err != nil {
		return
	}

	switch resp.StatusCode {
	case http.StatusOK:
		response, err = ioutil.ReadAll(resp.Body)
	case http.StatusNoContent:
		err = nil
	case http.StatusForbidden, http.StatusUnauthorized:
		// we're authenticated, but still got forbidden.
		// force password login to get a new token.
		client.Authenticator.Reset()
		err = errors.New(resp.Status)
	default:
		err = errors.New(resp.Status)
	}
	_ = resp.Body.Close()

	return
}

func (client *APIClient) buildRequest(ctx context.Context, method string, path string, payload string) (req *http.Request, err error) {
	req, _ = http.NewRequestWithContext(ctx, method, path, bytes.NewBufferString(payload))
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	var authHeaders http.Header
	authHeaders, err = client.Authenticator.AuthHeaders(ctx)
	if err != nil {
		return nil, fmt.Errorf("tado authentication failed: %s", err)
	}

	for key, values := range authHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return
}
