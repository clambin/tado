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
	}{
		{
			name:   "auto",
			config: ZoneAwayConfiguration{Type: "HEATING", AutoAdjust: true, ComfortLevel: Eco},
		},
		{
			name:   "manual",
			config: ZoneAwayConfiguration{Type: "HEATING", Setting: ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: Temperature{Celsius: 15.0}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, s := makeTestServer(tt.config, nil)
			defer s.Close()
			output, err := c.GetZoneAutoConfiguration(context.Background(), 1)
			require.NoError(t, err)
			assert.Equal(t, output, tt.config)
		})
	}
}

func TestAPIClient_SetZoneAwayAutoAdjust(t *testing.T) {
	tests := []struct {
		name         string
		comfortLevel ComfortLevel
		pass         bool
	}{
		{name: "eco", comfortLevel: Eco, pass: true},
		{name: "balance", comfortLevel: Balance, pass: true},
		{name: "comfort", comfortLevel: Comfort, pass: true},
		{name: "invalid", comfortLevel: 12, pass: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := zoneAwayHandler{}
			s := httptest.NewServer(http.HandlerFunc(h.Handle))
			defer s.Close()
			a := fakeAuthenticator{}
			c := newWithAuthenticator(&a)
			c.apiURL = buildURLMap(s.URL)

			err := c.SetZoneAwayAutoAdjust(context.Background(), 1, tt.comfortLevel)
			if !tt.pass {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			cfg, err := c.GetZoneAutoConfiguration(context.Background(), 1)
			require.NoError(t, err)
			assert.True(t, cfg.AutoAdjust)
			assert.Equal(t, tt.comfortLevel, cfg.ComfortLevel)
		})
	}
}

func TestAPIClient_SetZoneAwayManual(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		pass        bool
	}{
		{name: "on", temperature: 18, pass: true},
		{name: "off", temperature: 5, pass: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := zoneAwayHandler{}
			s := httptest.NewServer(http.HandlerFunc(h.Handle))
			defer s.Close()
			a := fakeAuthenticator{}
			c := newWithAuthenticator(&a)
			c.apiURL = buildURLMap(s.URL)

			err := c.SetZoneAwayManual(context.Background(), 1, tt.temperature)
			if !tt.pass {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			cfg, err := c.GetZoneAutoConfiguration(context.Background(), 1)
			assert.False(t, cfg.AutoAdjust)
			if tt.temperature <= 5.0 {
				assert.Equal(t, "OFF", cfg.Setting.Power)
			} else {
				assert.Equal(t, "ON", cfg.Setting.Power)
				assert.Equal(t, tt.temperature, cfg.Setting.Temperature.Celsius)
			}
		})
	}
}

type zoneAwayHandler struct {
	zoneConfig ZoneAwayConfiguration
}

func (h *zoneAwayHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/me":
		_, _ = w.Write([]byte(`{ "homes": [ { "id" : 242, "name": "home" } ] }`))
	default:
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(h.zoneConfig)
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&h.zoneConfig); err != nil {
				http.Error(w, "", http.StatusUnprocessableEntity)
			}
		default:
			http.Error(w, r.Method, http.StatusMethodNotAllowed)
		}

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

func TestAPIClient_GetZoneDayReport(t *testing.T) {
	var info DayReport

	c, s := makeTestServer(info, nil)
	defer s.Close()

	ctx := context.Background()
	output, err := c.GetZoneDayReport(ctx, 1, time.Now())
	require.NoError(t, err)
	assert.Equal(t, info, output)
}
