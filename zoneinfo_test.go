package tado_test

import (
	"context"
	"encoding/json"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestZoneInfo_GetState(t *testing.T) {
	tests := []struct {
		name     string
		zoneInfo tado.ZoneInfo
		state    tado.ZoneState
	}{
		{
			name: "auto",
			zoneInfo: tado.ZoneInfo{
				Setting: tado.ZoneInfoSetting{
					Power:       "ON",
					Temperature: tado.Temperature{Celsius: 17.0},
				},
			},
			state: tado.ZoneStateAuto,
		},
		{
			name: "manual",
			zoneInfo: tado.ZoneInfo{
				Setting: tado.ZoneInfoSetting{
					Power:       "ON",
					Temperature: tado.Temperature{Celsius: 17.0},
				},
				Overlay: tado.ZoneInfoOverlay{
					Type: "MANUAL",
					Setting: tado.ZoneInfoOverlaySetting{
						Type:        "HEATING",
						Power:       "ON",
						Temperature: tado.Temperature{Celsius: 22.0},
					},
					Termination: tado.ZoneInfoOverlayTermination{
						Type: "MANUAL",
					},
				},
			},
			state: tado.ZoneStateManual,
		},
		{
			name: "manual w/ termination",
			zoneInfo: tado.ZoneInfo{
				Setting: tado.ZoneInfoSetting{
					Power:       "ON",
					Temperature: tado.Temperature{Celsius: 17.0},
				},
				Overlay: tado.ZoneInfoOverlay{
					Type: "MANUAL",
					Setting: tado.ZoneInfoOverlaySetting{
						Type:        "HEATING",
						Power:       "ON",
						Temperature: tado.Temperature{Celsius: 22.0},
					},
					Termination: tado.ZoneInfoOverlayTermination{
						Type: "AUTO",
					},
				},
			},
			state: tado.ZoneStateTemporaryManual,
		},
		{
			name: "off",
			zoneInfo: tado.ZoneInfo{
				Setting: tado.ZoneInfoSetting{
					Power:       "OFF",
					Temperature: tado.Temperature{Celsius: 5.0},
				},
			},
			state: tado.ZoneStateOff,
		},
		{
			name: "manual off",
			zoneInfo: tado.ZoneInfo{
				Setting: tado.ZoneInfoSetting{
					Power:       "ON",
					Temperature: tado.Temperature{Celsius: 17.0},
				},
				Overlay: tado.ZoneInfoOverlay{
					Type: "MANUAL",
					Setting: tado.ZoneInfoOverlaySetting{
						Type:        "HEATING",
						Power:       "ON",
						Temperature: tado.Temperature{Celsius: 5.0},
					},
					Termination: tado.ZoneInfoOverlayTermination{
						Type: "AUTO",
					},
				},
			},
			state: tado.ZoneStateOff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.state, tt.zoneInfo.GetState())
		})
	}
}

func TestAPIClient_GetZoneInfo(t *testing.T) {
	response := tado.ZoneInfo{
		Setting: tado.ZoneInfoSetting{
			Power:       "ON",
			Temperature: tado.Temperature{Celsius: 19.0},
		},
		ActivityDataPoints: tado.ZoneInfoActivityDataPoints{HeatingPower: tado.Percentage{Percentage: 75.0}},
		SensorDataPoints: tado.ZoneInfoSensorDataPoints{
			Temperature: tado.Temperature{Celsius: 20.0},
			Humidity:    tado.Percentage{Percentage: 10.5},
		},
		OpenWindow: tado.ZoneInfoOpenWindow{},
		Overlay:    tado.ZoneInfoOverlay{},
	}

	c, s := makeTestServer(response)
	ctx := context.Background()
	zoneInfo, err := c.GetZoneInfo(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, response, zoneInfo)

	s.Close()
	_, err = c.GetZoneInfo(ctx, 1)
	assert.Error(t, err)

}

func TestAPIClient_ZoneOverlay(t *testing.T) {
	tests := []struct {
		name          string
		action        func(ctx context.Context, client *tado.APIClient) error
		expectedState tado.ZoneState
	}{
		{
			name: "manual",
			action: func(ctx context.Context, client *tado.APIClient) error {
				return client.SetZoneOverlay(ctx, 1, 18.0)
			},
			expectedState: tado.ZoneStateManual,
		},
		{
			name: "off",
			action: func(ctx context.Context, client *tado.APIClient) error {
				return client.SetZoneOverlay(ctx, 1, 1.0)
			},
			expectedState: tado.ZoneStateOff,
		},
		{
			name: "temp manual",
			action: func(ctx context.Context, client *tado.APIClient) error {
				return client.SetZoneOverlayWithDuration(ctx, 1, 18.0, time.Hour)
			},
			expectedState: tado.ZoneStateTemporaryManual,
		},
		{
			name: "temp off",
			action: func(ctx context.Context, client *tado.APIClient) error {
				return client.SetZoneOverlayWithDuration(ctx, 1, 1.0, time.Hour)
			},
			expectedState: tado.ZoneStateOff,
		},
		{
			name: "temp not temp",
			action: func(ctx context.Context, client *tado.APIClient) error {
				return client.SetZoneOverlayWithDuration(ctx, 1, 18.0, 0)
			},
			expectedState: tado.ZoneStateManual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newOverlayManager()
			s := httptest.NewServer(mgr)

			c := tado.New("", "", "")
			c.APIURL = s.URL
			c.Authenticator = &fakeAuthenticator{Token: "1234"}

			ctx := context.TODO()

			zoneInfo, err := c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, tado.ZoneStateAuto, zoneInfo.GetState())

			err = tt.action(ctx, c)
			require.NoError(t, err)

			zoneInfo, err = c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedState, zoneInfo.GetState())

			err = c.DeleteZoneOverlay(ctx, 1)
			require.NoError(t, err)

			zoneInfo, err = c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, tado.ZoneStateAuto, zoneInfo.GetState())

			s.Close()
			err = tt.action(ctx, c)
			assert.Error(t, err)
		})
	}
}

type overlayManager struct {
	zoneInfo tado.ZoneInfo
}

func newOverlayManager() *overlayManager {
	return &overlayManager{
		zoneInfo: tado.ZoneInfo{
			Setting:            tado.ZoneInfoSetting{Power: "ON", Temperature: tado.Temperature{Celsius: 22.5}},
			ActivityDataPoints: tado.ZoneInfoActivityDataPoints{HeatingPower: tado.Percentage{Percentage: 80.0}},
			SensorDataPoints:   tado.ZoneInfoSensorDataPoints{Temperature: tado.Temperature{Celsius: 20.0}, Humidity: tado.Percentage{Percentage: 75.0}},
		},
	}
}

func (o *overlayManager) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		if err := json.NewEncoder(w).Encode(o.zoneInfo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		if err := json.NewDecoder(req.Body).Decode(&o.zoneInfo.Overlay); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		o.zoneInfo.Overlay = tado.ZoneInfoOverlay{}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func TestZoneState_String(t *testing.T) {
	tests := []struct {
		name string
		s    tado.ZoneState
		want string
	}{
		{name: "ZoneStateUnknown", s: tado.ZoneStateUnknown, want: "unknown"},
		{name: "ZoneStateOff", s: tado.ZoneStateOff, want: "off"},
		{name: "ZoneStateAuto", s: tado.ZoneStateAuto, want: "auto"},
		{name: "ZoneStateTemporaryManual", s: tado.ZoneStateTemporaryManual, want: "manual (temp)"},
		{name: "ZoneStateManual", s: tado.ZoneStateManual, want: "manual"},
		{name: "invalid", s: tado.ZoneState(-1), want: "(invalid)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.String(), "String()")
		})
	}
}
