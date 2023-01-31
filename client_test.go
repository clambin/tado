package tado_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/clambin/tado"
	"github.com/clambin/tado/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIClient_Authentication(t *testing.T) {
	server := APIServer{}
	apiServer := httptest.NewServer(http.HandlerFunc(server.apiHandler))
	defer apiServer.Close()
	authenticator := mocks.NewAuthenticator(t)
	authenticator.
		On("AuthHeaders", mock.AnythingOfType("*context.emptyCtx")).
		Return(http.Header{"Authorization": []string{"Bearer bad_token"}}, nil).Once()
	authenticator.On("Reset").Once()

	client := tado.New("user@examle.com", "some-password", "")
	client.APIURL = apiServer.URL
	client.Authenticator = authenticator

	_, err := client.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "403 Forbidden", err.Error())

	authenticator.
		On("AuthHeaders", mock.AnythingOfType("*context.emptyCtx")).
		Return(http.Header{}, fmt.Errorf("server is down")).Once()

	_, err = client.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "tado authentication failed: server is down", err.Error())

	authenticator.
		On("AuthHeaders", mock.AnythingOfType("*context.emptyCtx")).
		Return(http.Header{"Authorization": []string{"Bearer good_token"}}, nil)

	_, err = client.GetZones(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 242, client.HomeID)

	server.fail = true
	_, err = client.GetZones(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestAPIClient_DecodeError(t *testing.T) {
	info := tado.MobileDevice{
		ID:   1,
		Name: "foo",
		Settings: tado.MobileDeviceSettings{
			GeoTrackingEnabled: false,
		},
		Location: tado.MobileDeviceLocation{},
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
	authenticator := mocks.NewAuthenticator(t)
	authenticator.
		On("AuthHeaders", mock.AnythingOfType("*context.timerCtx")).
		Return(http.Header{"Authorization": []string{"Bearer good_token"}}, nil)

	client := tado.New("user@examle.com", "some-password", "")
	client.APIURL = apiServer.URL
	client.Authenticator = authenticator

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := client.GetWeatherInfo(ctx)

	assert.Error(t, err)
}

func makeTestServer(response any) (*tado.APIClient, *httptest.Server) {
	const token = "1234"
	s := httptest.NewServer(authenticationHandler(token)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/me":
			_, _ = w.Write([]byte(`{ "homeId": 1 }`))
		default:
			_ = json.NewEncoder(w).Encode(response)
		}
	})))

	c := tado.New("", "", "")
	c.APIURL = s.URL
	c.Authenticator = fakeAuthenticator{Token: token}

	return c, s
}

type fakeAuthenticator struct {
	Token string
}

func (f fakeAuthenticator) AuthHeaders(_ context.Context) (header http.Header, err error) {
	return http.Header{"Authorization": []string{"Bearer " + f.Token}}, nil
}

func (f fakeAuthenticator) Reset() {
}

var _ tado.Authenticator = &fakeAuthenticator{}

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
