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

	c, _ := New(username, password, "")
	ctx := context.Background()
	zones, err := c.GetZones(ctx)
	require.NoError(t, err)

	for _, zone := range zones {
		_, err = c.GetZoneInfo(ctx, zone.ID)
		require.NoError(t, err)
	}
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
	c, s := makeTestServer(WeatherInfo{}, func(r *http.Request) bool { return wait(r.Context(), 5*time.Second) })
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := c.GetWeatherInfo(ctx)
	assert.Error(t, err)
}

func wait(ctx context.Context, duration time.Duration) (passed bool) {
	select {
	case <-time.After(duration):
		return true
	case <-ctx.Done():
		return false
	}
}

func TestAPIClient_TooManyRequests(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "slow down", http.StatusTooManyRequests)
	}))
	defer s.Close()

	var c APIClient
	c.HTTPClient = http.DefaultClient
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

	var c APIClient
	c.HTTPClient = http.DefaultClient
	c.apiURL = buildURLMap(s.URL)

	_, err := c.GetZones(context.Background())
	require.Error(t, err)
	assert.Equal(t, `unprocessable entity: {"foo":"bar"}`, err.Error())
}

func TestAPIClient_NoHomes(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{ "homes": [ ] }`))
	}))
	defer s.Close()

	var c APIClient
	c.HTTPClient = http.DefaultClient
	c.apiURL = buildURLMap(s.URL)

	_, err := c.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "no homes detected", err.Error())
}

func makeTestServer[T any](response T, f func(r *http.Request) bool) (*APIClient, *httptest.Server) {
	ts := testServer[T]{content: response}
	handler := http.Handler(&ts)
	if f != nil {
		next := handler
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !f(r) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	h := httptest.NewServer(handler)

	var c APIClient
	c.HTTPClient = http.DefaultClient
	c.apiURL = buildURLMap(h.URL)

	return &c, h
}

type testServer[T any] struct {
	content T
}

func (s *testServer[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/me":
		_, _ = w.Write([]byte(`{ "homes": [ { "id" : 242, "name": "home" } ] }`))
	default:
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(s.content)
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&s.content); err != nil {
				e := UnprocessableEntryError{err: &APIError{
					Errors: []errorEntry{{
						Code:  "unprocessable entry",
						Title: err.Error(),
					}},
				}}
				_, _ = w.Write([]byte(e.Error()))
				w.WriteHeader(http.StatusUnprocessableEntity)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
