package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetEnergySavings(t *testing.T) {
	info := struct {
		Reports []EnergySavingsReport
	}{Reports: []EnergySavingsReport{
		{
			YearMonth: "2013-01",
		},
	},
	}
	c, s := makeTestServer(info, nil)
	rcvd, err := c.GetEnergySavings(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info.Reports, rcvd)

	s.Close()
	_, err = c.GetEnergySavings(context.Background())
	assert.Error(t, err)

}
