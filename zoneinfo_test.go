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

func TestAPIClient_GetZoneCapabilities(t *testing.T) {
	info := ZoneCapabilities{Type: "HEATING"}
	info.Temperatures.Celsius.Min = 5
	info.Temperatures.Celsius.Max = 25
	info.Temperatures.Celsius.Step = 0.1

	c, s := makeTestServer(info, nil)
	ctx := context.Background()
	capabilities, err := c.GetZoneCapabilities(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, info, capabilities)

	s.Close()
	_, err = c.GetZoneCapabilities(ctx, 1)
	assert.Error(t, err)
}

func TestAPIClient_GetZoneEarlyStart(t *testing.T) {
	info := struct {
		Enabled bool
	}{Enabled: true}

	c, s := makeTestServer(info, nil)
	ctx := context.Background()
	earlyStart, err := c.GetZoneEarlyStart(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, info.Enabled, earlyStart)

	s.Close()
	_, err = c.GetZoneEarlyStart(ctx, 1)
	assert.Error(t, err)
}

func TestAPIClient_SetZoneEarlyStart(t *testing.T) {
	info := struct {
		Enabled bool
	}{Enabled: true}

	c, s := makeTestServer(info, nil)
	ctx := context.Background()
	err := c.SetZoneEarlyStart(ctx, 1, true)
	require.NoError(t, err)
	err = c.SetZoneEarlyStart(ctx, 1, false)
	require.NoError(t, err)

	s.Close()
	err = c.SetZoneEarlyStart(ctx, 1, true)
	assert.Error(t, err)
}

func TestAPIClient_GetZoneAutoConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config ZoneAwayConfiguration
		pass   bool
	}{
		{
			name:   "auto",
			config: ZoneAwayConfiguration{Type: "HEATING", AutoAdjust: true, ComfortLevel: Eco},
			pass:   true,
		},
		{
			name:   "manual",
			config: ZoneAwayConfiguration{Type: "HEATING", Setting: &ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: Temperature{Celsius: 15.0}}},
			pass:   true,
		},
		{
			name:   "invalid",
			config: ZoneAwayConfiguration{Type: "HEATING", AutoAdjust: true, ComfortLevel: 33},
			pass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, s := makeTestServer(tt.config, nil)
			output, err := c.GetZoneAutoConfiguration(context.Background(), 1)
			require.NoError(t, err)
			assert.Equal(t, output, tt.config)

			err = c.SetZoneAutoConfiguration(context.Background(), 1, tt.config)
			if !tt.pass {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			s.Close()
		})
	}
}
