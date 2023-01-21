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

/*
	func TestTypesToString(t *testing.T) {
		zoneInfo := tado.ZoneInfo{
			Setting: tado.ZoneInfoSetting{
				Power:       "ON",
				Temperature: tado.Temperature{Celsius: 25.0},
			},
			OpenWindow: tado.ZoneInfoOpenWindow{
				DurationInSeconds:      900,
				RemainingTimeInSeconds: 250,
			},
			SensorDataPoints: tado.ZoneInfoSensorDataPoints{
				Temperature: tado.Temperature{Celsius: 21.0},
				Humidity:    tado.Percentage{Percentage: 30.0},
			},
			ActivityDataPoints: tado.ZoneInfoActivityDataPoints{
				HeatingPower: tado.Percentage{Percentage: 25.0},
			},
			Overlay: tado.ZoneInfoOverlay{
				Type: "MANUAL",
				Setting: tado.ZoneInfoOverlaySetting{
					Type:        "HEATING",
					Power:       "ON",
					Temperature: tado.Temperature{Celsius: 25.0},
				},
				Termination: tado.ZoneInfoOverlayTermination{
					Type:          "TIMER",
					RemainingTime: 120,
				},
			},
		}

		assert.Equal(t, `target=25.0ºC, temp=21.0ºC, humidity=30.0%, heating=25.0%, power=ON, openwindow=650s, overlay={type=MANUAL, settings={type=HEATING, power=ON, temp=25.0ºC}, termination={type="TIMER", remaining=120}}`, zoneInfo.String())

		weatherInfo := tado.WeatherInfo{
			OutsideTemperature: tado.Temperature{Celsius: 27.0},
			SolarIntensity:     tado.Percentage{Percentage: 75.0},
			WeatherState:       tado.Value{Value: "SUNNY"},
		}

		assert.Equal(t, `temp=27.0ºC, solar=75.0%, weather=SUNNY`, weatherInfo.String())

		zone := tado.Zone{
			ID:   1,
			Name: "Living room",
			Devices: []tado.Device{
				{
					DeviceType:      "RU02",
					Firmware:        "67.2",
					ConnectionState: tado.ConnectionState{Value: true},
					BatteryState:    "LOW",
				},
			},
		}

		assert.Equal(t, "id=1 name=Living room devices={type=RU02 firmware=67.2 connection=true battery=LOW}", zone.String())

		mobileDevice := tado.MobileDevice{
			ID:   1,
			Name: "phone",
			Settings: tado.MobileDeviceSettings{
				GeoTrackingEnabled: true,
			},
			Location: tado.MobileDeviceLocation{
				Stale:  false,
				AtHome: true,
			},
		}

		assert.Equal(t, `name=phone, geotrack=true, stale=false, athome=true`, mobileDevice.String())
	}
*/
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
