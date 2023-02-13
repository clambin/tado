package tado

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIClient_GetZoneInfo_E2E(t *testing.T) {
	username := os.Getenv("TADO_USERNAME")
	password := os.Getenv("TADO_PASSWORD")

	if username == "" || password == "" {
		t.Skip("environment not set. skipping ...")
	}

	c := New(username, password, "")
	ctx := context.Background()
	zones, err := c.GetZones(ctx)
	require.NoError(t, err)

	for _, zone := range zones {
		zoneInfo, err := c.GetZoneInfo(ctx, zone.ID)
		require.NoError(t, err)
		t.Logf("%s: %s", zone.Name, zoneInfo.GetState())
	}
}

func TestAPIClient_Authentication(t *testing.T) {
	response := []Zone{
		{ID: 1, Name: "foo", Devices: []Device{{DeviceType: "foo", CurrentFwVersion: "v1.0", ConnectionState: ConnectionState{Value: true}, BatteryState: "OK"}}},
		{ID: 2, Name: "bar", Devices: []Device{{DeviceType: "bar", CurrentFwVersion: "v1.0", ConnectionState: ConnectionState{Value: false}, BatteryState: "OK"}}},
	}

	c, s := makeTestServer(response, nil)

	auth := fakeAuthenticator{}
	c.authenticator = &auth

	auth.Token = "4321"
	_, err := c.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "403 Forbidden", err.Error())

	auth.Token = "1234"
	_, err = c.GetZones(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 242, c.activeHomeID)

	s.Close()
	_, err = c.GetZones(context.Background())
	assert.Error(t, err)
}

func TestAPIClient_DecodeError(t *testing.T) {
	info := MobileDevice{
		ID:   1,
		Name: "foo",
		Settings: MobileDeviceSettings{
			GeoTrackingEnabled: false,
		},
		Location: MobileDeviceLocation{},
	}

	c, s := makeTestServer(info, nil)
	defer s.Close()

	_, err := c.GetMobileDevices(context.Background())
	assert.Error(t, err)
}

func TestAPIClient_Timeout(t *testing.T) {
	c, s := makeTestServer(nil, func(ctx context.Context) bool { return wait(ctx, 5*time.Second) })
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := c.GetWeatherInfo(ctx)
	assert.Error(t, err)
}

func wait(ctx context.Context, duration time.Duration) (passed bool) {
	timer := time.NewTimer(duration)
loop:
	for {
		select {
		case <-timer.C:
			break loop
		case <-ctx.Done():
			return false
		}
	}
	return true
}

func TestAPIClient_NoHomes(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{ "homes": [ ] }`))
	}))
	defer s.Close()

	c := New("", "", "")
	c.apiURL = buildURLMap(s.URL)
	c.authenticator = fakeAuthenticator{Token: "1234"}

	_, err := c.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "no homes detected", err.Error())
}

func makeTestServer(response any, middleware func(ctx context.Context) bool) (*APIClient, *httptest.Server) {
	const token = "1234"
	s := httptest.NewServer(authenticationHandler(token)(responder(response, middleware)))

	c := New("", "", "")
	c.apiURL = buildURLMap(s.URL)
	c.authenticator = fakeAuthenticator{Token: token}

	return c, s
}

func responder(response any, middleware func(ctx context.Context) bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if middleware != nil && !middleware(ctx) {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		switch r.URL.Path {
		case "/me":
			_, _ = w.Write([]byte(`{ "homes": [ { "id" : 242, "name": "home" } ] }`))
		default:
			_ = json.NewEncoder(w).Encode(response)
		}
	}
}

type fakeAuthenticator struct {
	Token string
}

func (f fakeAuthenticator) GetAuthToken(_ context.Context) (string, error) {
	return f.Token, nil
}

func (f fakeAuthenticator) Reset() {
}

var _ authenticator = &fakeAuthenticator{}

func authenticationHandler(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			if bearer != "Bearer "+token {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
