package tado

import (
	"context"
	"encoding/json"
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
		zoneInfo ZoneInfo
		state    ZoneState
	}{
		{
			name: "auto",
			zoneInfo: ZoneInfo{
				Setting: ZoneInfoSetting{
					Power:       "ON",
					Temperature: Temperature{Celsius: 17.0},
				},
			},
			state: ZoneStateAuto,
		},
		{
			name: "manual",
			zoneInfo: ZoneInfo{
				Setting: ZoneInfoSetting{
					Power:       "ON",
					Temperature: Temperature{Celsius: 17.0},
				},
				Overlay: ZoneInfoOverlay{
					Type: "MANUAL",
					Setting: ZoneInfoOverlaySetting{
						Type:        "HEATING",
						Power:       "ON",
						Temperature: Temperature{Celsius: 22.0},
					},
					Termination: ZoneInfoOverlayTermination{
						Type: "MANUAL",
					},
				},
			},
			state: ZoneStateManual,
		},
		{
			name: "manual w/ termination",
			zoneInfo: ZoneInfo{
				Setting: ZoneInfoSetting{
					Power:       "ON",
					Temperature: Temperature{Celsius: 17.0},
				},
				Overlay: ZoneInfoOverlay{
					Type: "MANUAL",
					Setting: ZoneInfoOverlaySetting{
						Type:        "HEATING",
						Power:       "ON",
						Temperature: Temperature{Celsius: 22.0},
					},
					Termination: ZoneInfoOverlayTermination{
						Type: "AUTO",
					},
				},
			},
			state: ZoneStateTemporaryManual,
		},
		{
			name: "off",
			zoneInfo: ZoneInfo{
				Setting: ZoneInfoSetting{
					Power:       "OFF",
					Temperature: Temperature{Celsius: 5.0},
				},
			},
			state: ZoneStateOff,
		},
		{
			name: "manual off",
			zoneInfo: ZoneInfo{
				Setting: ZoneInfoSetting{
					Power:       "ON",
					Temperature: Temperature{Celsius: 17.0},
				},
				Overlay: ZoneInfoOverlay{
					Type: "MANUAL",
					Setting: ZoneInfoOverlaySetting{
						Type:        "HEATING",
						Power:       "ON",
						Temperature: Temperature{Celsius: 5.0},
					},
					Termination: ZoneInfoOverlayTermination{
						Type: "AUTO",
					},
				},
			},
			state: ZoneStateOff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.state, tt.zoneInfo.GetState())
		})
	}
}

func TestAPIClient_GetZoneInfo(t *testing.T) {
	response := ZoneInfo{
		Setting: ZoneInfoSetting{
			Power:       "ON",
			Temperature: Temperature{Celsius: 19.0},
		},
		ActivityDataPoints: ZoneInfoActivityDataPoints{HeatingPower: Percentage{Percentage: 75.0}},
		SensorDataPoints: ZoneInfoSensorDataPoints{
			Temperature: Temperature{Celsius: 20.0},
			Humidity:    Percentage{Percentage: 10.5},
		},
		OpenWindow: ZoneInfoOpenWindow{},
		Overlay:    ZoneInfoOverlay{},
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
		action        func(ctx context.Context, client *APIClient) error
		expectedState ZoneState
	}{
		{
			name: "manual",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneOverlay(ctx, 1, 18.0)
			},
			expectedState: ZoneStateManual,
		},
		{
			name: "off",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneOverlay(ctx, 1, 1.0)
			},
			expectedState: ZoneStateOff,
		},
		{
			name: "temp manual",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneOverlayWithDuration(ctx, 1, 18.0, time.Hour)
			},
			expectedState: ZoneStateTemporaryManual,
		},
		{
			name: "temp off",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneOverlayWithDuration(ctx, 1, 1.0, time.Hour)
			},
			expectedState: ZoneStateOff,
		},
		{
			name: "temp not temp",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneOverlayWithDuration(ctx, 1, 18.0, 0)
			},
			expectedState: ZoneStateManual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newOverlayManager()
			s := httptest.NewServer(mgr)

			c := New("", "", "")
			c.apiURL = s.URL
			c.authenticator = &fakeAuthenticator{Token: "1234"}

			ctx := context.TODO()

			zoneInfo, err := c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, ZoneStateAuto, zoneInfo.GetState())

			err = tt.action(ctx, c)
			require.NoError(t, err)

			zoneInfo, err = c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedState, zoneInfo.GetState())

			err = c.DeleteZoneOverlay(ctx, 1)
			require.NoError(t, err)

			zoneInfo, err = c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, ZoneStateAuto, zoneInfo.GetState())

			s.Close()
			err = tt.action(ctx, c)
			assert.Error(t, err)
		})
	}
}

type overlayManager struct {
	zoneInfo ZoneInfo
}

func newOverlayManager() *overlayManager {
	return &overlayManager{
		zoneInfo: ZoneInfo{
			Setting:            ZoneInfoSetting{Power: "ON", Temperature: Temperature{Celsius: 22.5}},
			ActivityDataPoints: ZoneInfoActivityDataPoints{HeatingPower: Percentage{Percentage: 80.0}},
			SensorDataPoints:   ZoneInfoSensorDataPoints{Temperature: Temperature{Celsius: 20.0}, Humidity: Percentage{Percentage: 75.0}},
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
		o.zoneInfo.Overlay = ZoneInfoOverlay{}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func TestZoneState_String(t *testing.T) {
	tests := []struct {
		name string
		s    ZoneState
		want string
	}{
		{name: "ZoneStateUnknown", s: ZoneStateUnknown, want: "unknown"},
		{name: "ZoneStateOff", s: ZoneStateOff, want: "off"},
		{name: "ZoneStateAuto", s: ZoneStateAuto, want: "auto"},
		{name: "ZoneStateTemporaryManual", s: ZoneStateTemporaryManual, want: "manual (temp)"},
		{name: "ZoneStateManual", s: ZoneStateManual, want: "manual"},
		{name: "invalid", s: ZoneState(-1), want: "(invalid)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.String(), "String()")
		})
	}
}
