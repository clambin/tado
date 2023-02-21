package tado

import (
	"context"
	"encoding/json"
	"errors"
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
		_, err := c.GetZoneInfo(ctx, zone.ID)
		require.NoError(t, err)
	}
}

func TestAPIClient_Authentication(t *testing.T) {
	response := []Zone{
		{ID: 1, Name: "foo", Devices: []Device{{DeviceType: "foo", CurrentFwVersion: "v1.0", ConnectionState: State{Value: true}, BatteryState: "OK"}}},
		{ID: 2, Name: "bar", Devices: []Device{{DeviceType: "bar", CurrentFwVersion: "v1.0", ConnectionState: State{Value: false}, BatteryState: "OK"}}},
	}

	_, s := makeTestServer(response, nil)

	auth := fakeAuthenticator{}
	c := newWithAuthenticator(&auth)
	c.apiURL = buildURLMap(s.URL)

	auth.Token = "4321"
	_, err := c.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "403 Forbidden", err.Error())
	assert.False(t, auth.set)

	auth.Token = "1234"
	_, err = c.GetZones(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 242, c.activeHomeID)
	assert.True(t, auth.set)

	auth.Token = "4321"
	_, err = c.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "403 Forbidden", err.Error())
	assert.False(t, auth.set)

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
	c, s := makeTestServer(WeatherInfo{}, func(ctx context.Context) bool { return wait(ctx, 5*time.Second) })
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

func TestAPIClient_TooManyRequests(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "slow down", http.StatusTooManyRequests)
	}))
	defer s.Close()

	auth := fakeAuthenticator{Token: "1234"}
	c := newWithAuthenticator(&auth)
	c.apiURL = buildURLMap(s.URL)

	_, err := c.GetZones(context.Background())
	require.Error(t, err)
	assert.Equal(t, "429 Too Many Requests", err.Error())
}

func TestAPIClient_UnprocessableEntity(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, `{"errors":[{"code": "foo", "title": "bar"}]}`, http.StatusUnprocessableEntity)
	}))
	defer s.Close()

	auth := fakeAuthenticator{Token: "1234"}
	c := newWithAuthenticator(&auth)
	c.apiURL = buildURLMap(s.URL)

	_, err := c.GetZones(context.Background())
	require.Error(t, err)
	assert.Equal(t, "unprocessable entity: bar", err.Error())
}

func TestAPIClient_NoHomes(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{ "homes": [ ] }`))
	}))
	defer s.Close()

	auth := fakeAuthenticator{Token: "1234"}
	c := newWithAuthenticator(&auth)
	c.apiURL = buildURLMap(s.URL)

	_, err := c.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "no homes detected", err.Error())
}

func makeTestServer[T any](response T, middleware func(ctx context.Context) bool) (*APIClient, *httptest.Server) {
	s := testServer[T]{content: response}
	const token = "1234"
	h := httptest.NewServer(authenticationHandler(token)(s.handler(middleware)))

	c := newWithAuthenticator(&fakeAuthenticator{Token: token})
	c.apiURL = buildURLMap(h.URL)

	return c, h
}

type testServer[T any] struct {
	content T
}

func (s *testServer[T]) handler(middleware func(ctx context.Context) bool) http.HandlerFunc {
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
			switch r.Method {
			case http.MethodGet:
				_ = json.NewEncoder(w).Encode(s.content)
			case http.MethodPut:
				if err := json.NewDecoder(r.Body).Decode(&s.content); err != nil {
					e := ErrUnprocessableEntry{Err: &Error{
						Errors: []errorEntry{{
							Code:  "unprocessable entry",
							Title: err.Error(),
						}},
					}}
					_ = json.NewEncoder(w).Encode(e)
					w.WriteHeader(http.StatusUnprocessableEntity)
				}
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}
}

type fakeAuthenticator struct {
	set   bool
	fail  bool
	Token string
}

func (f *fakeAuthenticator) GetAuthToken(_ context.Context) (string, error) {
	if f.fail {
		return "", errors.New("failed")
	}
	f.set = true
	return f.Token, nil
}

func (f *fakeAuthenticator) Reset() {
	f.set = false
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
