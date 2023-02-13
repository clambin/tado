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
				return client.SetZoneTemporaryOverlay(ctx, 1, 18.0, time.Hour)
			},
			expectedState: ZoneStateTemporaryManual,
		},
		{
			name: "temp off",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneTemporaryOverlay(ctx, 1, 1.0, time.Hour)
			},
			expectedState: ZoneStateOff,
		},
		{
			name: "temp not temp",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneTemporaryOverlay(ctx, 1, 18.0, 0)
			},
			expectedState: ZoneStateManual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newOverlayManager()
			s := httptest.NewServer(mgr)

			c := New("", "", "")
			c.apiURL = buildURLMap(s.URL)
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
		if req.URL.Path == "/me" {
			_, _ = w.Write([]byte(`{ "homes": [ { "id" : 1 } ] }`))
		} else if err := json.NewEncoder(w).Encode(o.zoneInfo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
