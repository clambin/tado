package tado

import (
	"context"
	"net/http"
)

// WeatherInfo contains the response to /api/v2/homes/<HomeID>/weather
//
// This structure provides the following key information:
//
//	OutsideTemperature.Celsius:  outside temperate, in degrees Celsius
//	SolarIntensity.Percentage:   solar intensity (0-100%)
//	WeatherState.Value:          string describing current weather (list TBD)
type WeatherInfo struct {
	OutsideTemperature Temperature `json:"outsideTemperature"`
	SolarIntensity     Percentage  `json:"solarIntensity"`
	WeatherState       Value       `json:"weatherState"`
}

// GetWeatherInfo retrieves weather information for the user's Home.
func (c *APIClient) GetWeatherInfo(ctx context.Context) (weatherInfo WeatherInfo, err error) {
	if err = c.initialize(ctx); err != nil {
		return
	}
	err = c.call(ctx, http.MethodGet, "myTado", "/weather", nil, &weatherInfo)
	return
}
