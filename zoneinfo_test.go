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
	configs := []ZoneAwayConfiguration{
		{Type: "HEATING", AutoAdjust: true, ComfortLevel: Eco},
		{Type: "HEATING", Setting: &ZonePowerSettings{Type: "HEATING", Power: "ON", Temperature: Temperature{Celsius: 15.0}}},
	}

	ctx := context.Background()
	for _, config := range configs {
		c, s := makeTestServer(config, nil)
		output, err := c.GetZoneAutoConfiguration(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, config, output)

		err = c.SetZoneAutoConfiguration(ctx, 1, config)
		assert.NoError(t, err)

		s.Close()
		_, err = c.GetZoneAutoConfiguration(ctx, 1)
		require.Error(t, err)
		err = c.SetZoneAutoConfiguration(ctx, 1, config)
		assert.Error(t, err)
	}
}

func TestAPIClient_GetZoneSchedule(t *testing.T) {
	schedules := []Schedule{
		{DayType: "MONDAY_TO_FRIDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSettings{Type: "HEATING", Power: "OFF"}},
		{DayType: "SATURDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSettings{Type: "HEATING", Power: "OFF"}},
		{DayType: "SUNDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSettings{Type: "HEATING", Power: "OFF"}},
	}

	c, s := makeTestServer(schedules, nil)
	defer s.Close()
	output, err := c.GetZoneSchedule(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, schedules, output)
}

func TestAPIClient_GetZoneScheduleForDay(t *testing.T) {
	schedules := []Schedule{
		{DayType: "MONDAY_TO_FRIDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSettings{Type: "HEATING", Power: "OFF"}},
	}

	c, s := makeTestServer(schedules, nil)
	defer s.Close()
	output, err := c.GetZoneScheduleForDay(context.Background(), 1, "MONDAY_TO_FRIDAY")
	require.NoError(t, err)
	assert.Equal(t, schedules, output)

	err = c.SetZoneScheduleForDay(context.Background(), 1, schedules[0].DayType, schedules)
	assert.NoError(t, err)
}
