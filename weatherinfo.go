package tado

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WeatherInfo contains the response to /api/v2/homes/<HomeID>/weather
//
// This structure provides the following key information:
//   OutsideTemperature.Celsius:  outside temperate, in degrees Celsius
//   SolarIntensity.Percentage:   solar intensity (0-100%)
//   WeatherState.Value:          string describing current weather (list TBD)
type WeatherInfo struct {
	OutsideTemperature Temperature `json:"outsideTemperature"`
	SolarIntensity     Percentage  `json:"solarIntensity"`
	WeatherState       Value       `json:"weatherState"`
}

// GetWeatherInfo retrieves weather information for the user's Home.
func (client *APIClient) GetWeatherInfo(ctx context.Context) (weatherInfo WeatherInfo, err error) {
	if err = client.initialize(ctx); err == nil {
		var body []byte
		if body, err = client.call(ctx, http.MethodGet, client.apiV2URL("/weather"), ""); err == nil {
			err = json.Unmarshal(body, &weatherInfo)
		}
	}
	return
}

// String converts WeatherInfo to a loggable string
func (weatherInfo WeatherInfo) String() string {
	return fmt.Sprintf("temp=%.1fºC, solar=%.1f%%, weather=%s",
		weatherInfo.OutsideTemperature.Celsius,
		weatherInfo.SolarIntensity.Percentage,
		weatherInfo.WeatherState.Value,
	)
}
