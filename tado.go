package tado

import (
	"bytes"
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
	GetHomes(context.Context) (Homes, error)
	SetActiveHome(context.Context, int) error
	SetActiveHomeByName(context.Context, string) error
	GetActiveHome(context.Context) (Home, bool)
	GetHomeInfo(context.Context) (HomeInfo, error)
	GetUsers(context.Context) ([]User, error)
	GetMobileDevices(context.Context) ([]MobileDevice, error)
	GetWeatherInfo(context.Context) (WeatherInfo, error)
	GetZones(context.Context) (Zones, error)
	GetZoneInfo(context.Context, int) (ZoneInfo, error)
	GetZoneCapabilities(context.Context, int) (ZoneCapabilities, error)
	GetZoneEarlyStart(context.Context, int) (bool, error)
	SetZoneEarlyStart(context.Context, int, bool) error
	GetZoneAutoConfiguration(context.Context, int) (ZoneAwayConfiguration, error)
	SetZoneAwayAutoAdjust(ctx context.Context, zoneID int, comfortLevel ComfortLevel) error
	SetZoneAwayManual(ctx context.Context, zoneID int, temperature float64) error
	SetZoneOverlay(context.Context, int, float64) error
	SetZoneTemporaryOverlay(context.Context, int, float64, time.Duration) error
	DeleteZoneOverlay(context.Context, int) error
	GetAirComfort(context.Context) (AirComfort, error)
	GetConsumption(context.Context, string, time.Time, time.Time) (Consumption, error)
	GetEnergySavings(context.Context) ([]EnergySavingsReport, error)
	GetRunningTimes(context.Context, time.Time, time.Time) ([]RunningTime, error)
	GetHeatingCircuits(context.Context) ([]HeatingCircuit, error)
	GetTimeTables(context.Context, int) ([]Timetable, error)
	GetActiveTimeTable(context.Context, int) (Timetable, error)
	SetActiveTimeTable(context.Context, int, Timetable) error
	GetTimeTableBlocks(context.Context, int, TimetableID) ([]Block, error)
	GetTimeTableBlocksForDayType(context.Context, int, TimetableID, string) ([]Block, error)
	SetTimeTableBlocksForDayType(context.Context, int, TimetableID, string, []Block) error
	GetHomeState(ctx context.Context) (homeState HomeState, err error)
	SetHomeState(ctx context.Context, home bool) error
	UnsetHomeState(ctx context.Context) error
	GetDefaultOverlay(ctx context.Context, zoneID int) (DefaultOverlay, error)
	SetDefaultOverlay(ctx context.Context, zoneID int, mode DefaultOverlay) error
	GetZoneMeasuringDevice(ctx context.Context, zoneID int) (measuringDevice ZoneMeasuringDevice, err error)
	GetZoneDayReport(ctx context.Context, zoneID int, date time.Time) (report DayReport, err error)
}

var _ API = &APIClient{}

// APIClient represents a Tado API client.
type APIClient struct {
	// HTTPClient is used to perform HTTP requests
	HTTPClient *http.Client

	authenticator
	apiURL       map[string]string
	lock         sync.RWMutex
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

	return newWithAuthenticator(&auth.Authenticator{
		HTTPClient:   http.DefaultClient,
		ClientID:     "tado-web-app",
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
		AuthURL:      "https://auth.tado.com/oauth/token",
	})
}

func newWithAuthenticator(auth authenticator) *APIClient {
	return &APIClient{
		HTTPClient: &http.Client{
			Transport: roundTripper{authenticator: auth},
		},
		authenticator: auth,
		apiURL:        buildURLMap(""),
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

// callAPI is implemented as a function rather than a method, because methods cannot have type parameters (yet?)
func callAPI[T any](ctx context.Context, c *APIClient, method, apiClass, endpoint string, request any) (response T, err error) {
	if apiClass != "me" {
		if err = c.setActiveHomeID(ctx); err != nil {
			return
		}
	}

	target := c.makeAPIURL(apiClass, endpoint)
	reqBody := new(bytes.Buffer)
	if request != nil {
		if err = json.NewEncoder(reqBody).Encode(request); err != nil {
			return response, fmt.Errorf("encode: %w", err)
		}
	}

	req, _ := http.NewRequestWithContext(ctx, method, target, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, fmt.Errorf("read: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength != 0 {
			err = json.Unmarshal(respBody, &response)
		}
	case http.StatusNoContent:
	case http.StatusForbidden, http.StatusUnauthorized:
		// we're authenticated, but still got forbidden.
		// force password login to get a new token.
		c.authenticator.Reset()
		err = errors.New(resp.Status)
	case http.StatusUnprocessableEntity:
		err = &UnprocessableEntryError{Err: parseError(respBody)}
	default:
		if err = parseError(respBody); !errors.Is(err, &Error{}) {
			err = errors.New(resp.Status)
		}
	}
	return
}

func parseError(body []byte) error {
	var errs Error
	if err := json.Unmarshal(body, &errs); err != nil {
		return fmt.Errorf("unparsable error: %w", err)
	}
	return &errs
}

func (c *APIClient) setActiveHomeID(ctx context.Context) (err error) {
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
