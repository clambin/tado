package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetWeatherInfo(t *testing.T) {
	info := WeatherInfo{
		OutsideTemperature: Temperature{Celsius: 18.5},
		SolarIntensity:     Percentage{Percentage: 64.0},
		WeatherState:       Value{Value: "SUN"},
	}

	c, s := makeTestServer(info, nil)
	rcvd, err := c.GetWeatherInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info, rcvd)

	s.Close()
	_, err = c.GetWeatherInfo(context.Background())
	assert.Error(t, err)
}
