package tado

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAPIClient_Authentication(t *testing.T) {
	server := APIServer{}
	apiServer := httptest.NewServer(http.HandlerFunc(server.apiHandler))
	defer apiServer.Close()
	authenticator := fakeAuthenticator{}

	client := New("user@examle.com", "some-password", "")
	client.apiURL = apiServer.URL
	client.authenticator = &authenticator

	authenticator.Token = "bad_token"
	_, err := client.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "403 Forbidden", err.Error())

	authenticator.Token = "good_token"
	_, err = client.GetZones(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 242, client.HomeID)

	server.fail = true
	_, err = client.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "500 Internal Server Error", err.Error())
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

	c, s := makeTestServer(info)
	_, err := c.GetMobileDevices(context.Background())
	assert.Error(t, err)
	s.Close()
}

func TestAPIClient_Timeout(t *testing.T) {
	server := APIServer{slow: true}
	apiServer := httptest.NewServer(http.HandlerFunc(server.apiHandler))
	defer apiServer.Close()
	authenticator := fakeAuthenticator{Token: "good_token"}

	client := New("user@example.com", "some-password", "")
	client.apiURL = apiServer.URL
	client.authenticator = authenticator

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := client.GetWeatherInfo(ctx)

	assert.Error(t, err)
}

func makeTestServer(response any) (*APIClient, *httptest.Server) {
	const token = "1234"
	s := httptest.NewServer(authenticationHandler(token)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/me":
			_, _ = w.Write([]byte(`{ "homeId": 1 }`))
		default:
			_ = json.NewEncoder(w).Encode(response)
		}
	})))

	c := New("", "", "")
	c.apiURL = s.URL
	c.authenticator = fakeAuthenticator{Token: token}

	return c, s
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
