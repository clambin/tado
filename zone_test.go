package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetZones(t *testing.T) {
	response := Zones{
		{ID: 1, Name: "foo", Devices: []Device{{DeviceType: "foo", CurrentFwVersion: "v1.0", ConnectionState: ConnectionState{Value: true}, BatteryState: "OK"}}},
		{ID: 2, Name: "bar", Devices: []Device{{DeviceType: "bar", CurrentFwVersion: "v1.0", ConnectionState: ConnectionState{Value: false}, BatteryState: "OK"}}},
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
