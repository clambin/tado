package tado

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/clambin/tado/auth"
)

// API for the Tado APIClient.
//
// Deprecated: will be removed in a future version. Clients should make their own, tailor-made mocks if needed.
//
//go:generate mockery --name API
type API interface {
	GetAccount(context.Context) (Account, error)
	GetHomes(context.Context) ([]string, error)
	SetActiveHome(context.Context, string) error
	GetActiveHome(context.Context) (Home, bool)
	GetHomeInfo(context.Context) (HomeInfo, error)
	GetUsers(context.Context) ([]User, error)
	GetMobileDevices(context.Context) ([]MobileDevice, error)
	GetWeatherInfo(context.Context) (WeatherInfo, error)
	GetZones(context.Context) ([]Zone, error)
	GetZoneInfo(context.Context, int) (ZoneInfo, error)
	GetZoneCapabilities(context.Context, int) (ZoneCapabilities, error)
	GetZoneEarlyStart(context.Context, int) (bool, error)
	SetZoneEarlyStart(context.Context, int, bool) error
	GetZoneAutoConfiguration(context.Context, int) (ZoneAwayConfiguration, error)
	SetZoneAutoConfiguration(context.Context, int, ZoneAwayConfiguration) error
	SetZoneOverlay(context.Context, int, float64) error
	SetZoneTemporaryOverlay(context.Context, int, float64, time.Duration) error
	DeleteZoneOverlay(context.Context, int) error
	GetAirComfort(context.Context) (AirComfort, error)
	GetConsumption(context.Context, string, time.Time, time.Time) (Consumption, error)
	GetEnergySavings(context.Context) ([]EnergySavingsReport, error)
	GetRunningTimes(context.Context, time.Time, time.Time) ([]RunningTime, error)
	GetHeatingCircuits(context.Context) ([]HeatingCircuit, error)
	GetTimeTables(context.Context, int) ([]TimeTable, error)
	GetActiveTimeTable(context.Context, int) (TimeTable, error)
	SetActiveTimeTable(context.Context, int, TimeTable) error
	GetTimeTableBlocks(context.Context, int, int) ([]Block, error)
	GetTimeTableBlocksForDayType(context.Context, int, int, string) ([]Block, error)
	SetTimeTableBlocksForDayType(context.Context, int, int, string, []Block) error
}

// APIClient represents a Tado API client.
type APIClient struct {
	// authenticator handles logging in to the Tado server
	authenticator
	// HTTPClient is used to perform HTTP requests
	HTTPClient *http.Client

	lock         sync.RWMutex
	apiURL       map[string]string
	account      *Account
	activeHomeID int
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
		apiURL:     buildURLMap(""),
	}
}

func buildURLMap(override string) map[string]string {
	myTado := "https://my.tado.com/api/v2"
	minder := "https://minder.tado.com/v1"
	bob := "https://energy-bob.tado.com"
	insights := "https://energy-insights.tado.com/api"

	if override != "" {
		myTado = override
		minder = override
		bob = override
		insights = override
	}

	return map[string]string{
		"me":       myTado + "/me",
		"myTado":   myTado + "/homes/%d",
		"minder":   minder + "/homes/%d",
		"bob":      bob + "/%d",
		"insights": insights + "/homes/%d",
	}
}

func (c *APIClient) initialize(ctx context.Context) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.activeHomeID > 0 {
		return
	}
	account, err := c.GetAccount(ctx)
	if err != nil {
		return err
	}
	c.account = &account
	if len(c.account.Homes) == 0 {
		return fmt.Errorf("no homes detected")
	}
	c.activeHomeID = c.account.Homes[0].ID
	return nil
}

func (c *APIClient) call(ctx context.Context, method string, apiClass, endpoint string, payload io.Reader, response any) error {
	req, err := c.buildRequest(ctx, method, c.makeAPIURL(apiClass, endpoint), payload)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
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
		c.authenticator.Reset()
		err = errors.New(resp.Status)
	default:
		err = errors.New(resp.Status)
	}

	return err
}

func (c *APIClient) makeAPIURL(apiClass string, endpoint string) string {
	base, ok := c.apiURL[apiClass]
	if !ok {
		panic("invalid api selector: " + base)
	}
	if apiClass == "me" {
		return base
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	return fmt.Sprintf(base, c.activeHomeID) + endpoint
}

func (c *APIClient) buildRequest(ctx context.Context, method string, path string, payload io.Reader) (*http.Request, error) {
	var req *http.Request
	token, err := c.authenticator.GetAuthToken(ctx)
	if err == nil {
		req, _ = http.NewRequestWithContext(ctx, method, path, payload)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		err = fmt.Errorf("auth: %w", err)
	}

	return req, err
}
