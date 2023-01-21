package tado_test

import (
	"context"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetZones(t *testing.T) {
	response := []tado.Zone{
		{ID: 1, Name: "foo", Devices: []tado.Device{{DeviceType: "foo", Firmware: "v1.0", ConnectionState: tado.ConnectionState{Value: true}, BatteryState: "OK"}}},
		{ID: 2, Name: "bar", Devices: []tado.Device{{DeviceType: "bar", Firmware: "v1.0", ConnectionState: tado.ConnectionState{Value: false}, BatteryState: "OK"}}},
	}

	c, s := makeTestServer(response)
	ctx := context.Background()
	zones, err := c.GetZones(ctx)
	require.NoError(t, err)
	assert.Equal(t, response, zones)

	s.Close()
	_, err = c.GetZones(ctx)
	assert.Error(t, err)
}
