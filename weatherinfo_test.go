package tado_test

import (
	"context"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetWeatherInfo(t *testing.T) {
	info := tado.WeatherInfo{
		OutsideTemperature: tado.Temperature{Celsius: 18.5},
		SolarIntensity:     tado.Percentage{Percentage: 64.0},
		WeatherState:       tado.Value{Value: "SUN"},
	}

	c, s := makeTestServer(info)
	rcvd, err := c.GetWeatherInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info, rcvd)

	s.Close()
	_, err = c.GetWeatherInfo(context.Background())
	assert.Error(t, err)
}
