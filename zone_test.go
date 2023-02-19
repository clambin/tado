package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetZones(t *testing.T) {
	response := Zones{
		{ID: 1, Name: "foo", Devices: []Device{{DeviceType: "foo", CurrentFwVersion: "v1.0", ConnectionState: State{Value: true}, BatteryState: "OK"}}},
		{ID: 2, Name: "bar", Devices: []Device{{DeviceType: "bar", CurrentFwVersion: "v1.0", ConnectionState: State{Value: false}, BatteryState: "OK"}}},
	}

	c, s := makeTestServer(response, nil)
	ctx := context.Background()
	zones, err := c.GetZones(ctx)
	require.NoError(t, err)
	assert.Equal(t, response, zones)

	s.Close()
	_, err = c.GetZones(ctx)
	assert.Error(t, err)
}

func TestZones_GetZone(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name  string
		z     Zones
		args  args
		want  Zone
		want1 bool
	}{
		{
			name:  "empty",
			z:     nil,
			args:  args{id: 1},
			want1: false,
		},
		{
			name:  "match",
			z:     Zones{Zone{ID: 1, Name: "foo"}, Zone{ID: 2, Name: "bar"}},
			args:  args{id: 1},
			want:  Zone{ID: 1, Name: "foo"},
			want1: true,
		},
		{
			name:  "mismatch",
			z:     Zones{Zone{ID: 1, Name: "foo"}, Zone{ID: 2, Name: "bar"}},
			args:  args{id: 3},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.z.GetZone(tt.args.id)
			assert.Equalf(t, tt.want, got, "GetZone(%v)", tt.args.id)
			assert.Equalf(t, tt.want1, got1, "GetZone(%v)", tt.args.id)
		})
	}
}

func TestZones_GetZoneByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name  string
		z     Zones
		args  args
		want  Zone
		want1 bool
	}{
		{
			name:  "empty",
			z:     nil,
			args:  args{name: "foo"},
			want1: false,
		},
		{
			name:  "match",
			z:     Zones{Zone{ID: 1, Name: "foo"}, Zone{ID: 2, Name: "bar"}},
			args:  args{name: "foo"},
			want:  Zone{ID: 1, Name: "foo"},
			want1: true,
		},
		{
			name:  "mismatch",
			z:     Zones{Zone{ID: 1, Name: "foo"}, Zone{ID: 2, Name: "bar"}},
			args:  args{name: "snafu"},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.z.GetZoneByName(tt.args.name)
			assert.Equalf(t, tt.want, got, "GetZone(%v)", tt.args.name)
			assert.Equalf(t, tt.want1, got1, "GetZone(%v)", tt.args.name)
		})
	}
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

func TestAPIClient_GetZoneMeasuringDevice(t *testing.T) {
	info := ZoneMeasuringDevice{
		BatteryState:     "TRUE",
		ConnectionState:  State{Value: true},
		CurrentFwVersion: "v1.0",
		SerialNo:         "123",
		ShortSerialNo:    "123",
	}

	c, s := makeTestServer(info, nil)
	defer s.Close()

	ctx := context.Background()
	output, err := c.GetZoneMeasuringDevice(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, info, output)

}
