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
		expectedState OverlayTerminationMode
	}{
		{
			name: "manual",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneOverlay(ctx, 1, 18.0)
			},
			expectedState: PermanentOverlay,
		},
		{
			name: "temp manual",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneTemporaryOverlay(ctx, 1, 18.0, time.Hour)
			},
			expectedState: TimerOverlay,
		},
		{
			name: "manual (duration not set)",
			action: func(ctx context.Context, client *APIClient) error {
				return client.SetZoneTemporaryOverlay(ctx, 1, 18.0, 0)
			},
			expectedState: PermanentOverlay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newOverlayManager()
			s := httptest.NewServer(mgr)

			auth := fakeAuthenticator{Token: "1234"}
			c := newWithAuthenticator(&auth)
			c.apiURL = buildURLMap(s.URL)

			ctx := context.TODO()

			zoneInfo, err := c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, NoOverlay, zoneInfo.Overlay.GetMode())

			err = tt.action(ctx, c)
			require.NoError(t, err)

			zoneInfo, err = c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedState, zoneInfo.Overlay.GetMode())

			err = c.DeleteZoneOverlay(ctx, 1)
			require.NoError(t, err)

			zoneInfo, err = c.GetZoneInfo(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, NoOverlay, zoneInfo.Overlay.GetMode())

			s.Close()
			err = tt.action(ctx, c)
			assert.Error(t, err)
		})
	}
}

type overlayManager struct {
	zoneInfo      ZoneInfo
	savedSettings *ZonePowerSetting
}

func newOverlayManager() *overlayManager {
	return &overlayManager{
		zoneInfo: ZoneInfo{
			Setting:            ZonePowerSetting{Power: "ON", Temperature: Temperature{Celsius: 22.5}},
			ActivityDataPoints: ZoneInfoActivityDataPoints{HeatingPower: Percentage{Percentage: 80.0}},
			SensorDataPoints:   ZoneInfoSensorDataPoints{InsideTemperature: Temperature{Celsius: 20.0}, Humidity: Percentage{Percentage: 75.0}},
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
		var overlay ZoneInfoOverlay
		if err := json.NewDecoder(req.Body).Decode(&overlay); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		overlay.Termination.TypeSkillBasedApp = overlay.Termination.Type
		if o.savedSettings == nil {
			o.savedSettings = &ZonePowerSetting{}
			*o.savedSettings = o.zoneInfo.Setting
		}
		o.zoneInfo.Overlay = overlay
		o.zoneInfo.Setting = overlay.Setting
	case http.MethodDelete:
		o.zoneInfo.Overlay = ZoneInfoOverlay{}
		if o.savedSettings != nil {
			o.zoneInfo.Setting = *o.savedSettings
		}
		o.savedSettings = nil
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func TestDefaultOverlay_IsValid(t *testing.T) {
	tests := []struct {
		name            string
		terminationType string
		terminationTime int
		reason          string
		valid           bool
	}{
		{
			name:            "manual",
			terminationType: "MANUAL",
			valid:           true,
		},
		{
			name:            "tado",
			terminationType: "TADO_MODE",
			valid:           true,
		},
		{
			name:            "timer",
			terminationType: "TIMER",
			terminationTime: 3600,
			valid:           true,
		},
		{
			name:            "invalid timer",
			terminationType: "TIMER",
			reason:          "DurationInSeconds must be set for TIMER overlay",
			valid:           false,
		},
		{
			name:            "invalid type",
			terminationType: "foo",
			reason:          "invalid type: foo",
			valid:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var o DefaultOverlay
			o.TerminationCondition.Type = tt.terminationType
			o.TerminationCondition.DurationInSeconds = tt.terminationTime

			reason, valid := o.isValid()
			assert.Equal(t, tt.reason, reason)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestAPIClient_GetDefaultOverlay(t *testing.T) {
	var response DefaultOverlay
	response.TerminationCondition.Type = "MANUAL"
	c, s := makeTestServer(response, nil)
	defer s.Close()

	overlay, err := c.GetDefaultOverlay(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "MANUAL", overlay.TerminationCondition.Type)
}

func TestAPIClient_SetDefaultOverlay(t *testing.T) {
	tests := []struct {
		name            string
		terminationType string
		terminationTime int
		valid           bool
	}{
		{
			name:            "manual",
			terminationType: "MANUAL",
			valid:           true,
		},
		{
			name:            "tado",
			terminationType: "TADO_MODE",
			valid:           true,
		},
		{
			name:            "timer",
			terminationType: "TIMER",
			terminationTime: 3600,
			valid:           true,
		},
		{
			name:            "invalid timer",
			terminationType: "TIMER",
			valid:           false,
		},
		{
			name:            "invalid type",
			terminationType: "foo",
			valid:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var o DefaultOverlay
			o.TerminationCondition.Type = tt.terminationType
			o.TerminationCondition.DurationInSeconds = tt.terminationTime

			c, s := makeTestServer(o, nil)
			defer s.Close()

			err := c.SetDefaultOverlay(context.Background(), 1, o)
			if !tt.valid {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
