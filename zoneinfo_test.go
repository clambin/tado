package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
					Setting: ZonePowerSetting{
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
					Setting: ZonePowerSetting{
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
					Setting: ZonePowerSetting{
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

	c, s := makeTestServer(response, nil)
	ctx := context.Background()
	zoneInfo, err := c.GetZoneInfo(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, response, zoneInfo)

	s.Close()
	_, err = c.GetZoneInfo(ctx, 1)
	assert.Error(t, err)
}
