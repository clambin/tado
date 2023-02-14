package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetAirComfort(t *testing.T) {
	info := AirComfort{
		Comfort: []ZoneAirComfort{
			{RoomId: 1, TemperatureLevel: "COLD", HumidityLevel: "COMFY"},
		},
	}

	c, s := makeTestServer(info, nil)
	rcvd, err := c.GetAirComfort(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info, rcvd)

	s.Close()
	_, err = c.GetAirComfort(context.Background())
	assert.Error(t, err)
}
