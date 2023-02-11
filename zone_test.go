package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetZones(t *testing.T) {
	response := []Zone{
		{ID: 1, Name: "foo", Devices: []Device{{DeviceType: "foo", Firmware: "v1.0", ConnectionState: ConnectionState{Value: true}, BatteryState: "OK"}}},
		{ID: 2, Name: "bar", Devices: []Device{{DeviceType: "bar", Firmware: "v1.0", ConnectionState: ConnectionState{Value: false}, BatteryState: "OK"}}},
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
