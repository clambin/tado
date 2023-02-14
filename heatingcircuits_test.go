package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetHeatingCircuits(t *testing.T) {
	info := []HeatingCircuit{
		{DriverSerialNo: "foo", DriverShortSerialNo: "foo", Number: 1},
		{DriverSerialNo: "bar", DriverShortSerialNo: "bar", Number: 1},
	}

	c, s := makeTestServer(info, nil)
	rcvd, err := c.GetHeatingCircuits(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info, rcvd)

	s.Close()
	_, err = c.GetHeatingCircuits(context.Background())
	assert.Error(t, err)
}
