package tado

import (
	"context"
	"net/http"
	"time"
)

// AirComfort contains the air comfort for a home. This contains the overall air comfort for the house, along with details for each zone.
type AirComfort struct {
	Freshness struct {
		Value          string    `json:"value"`
		LastOpenWindow time.Time `json:"lastOpenWindow"`
	} `json:"freshness"`
	Comfort []ZoneAirComfort `json:"comfort"`
}

// ZoneAirComfort contains the air comfort for one room / zone in the home
type ZoneAirComfort struct {
	RoomID           int    `json:"roomId"`
	TemperatureLevel string `json:"temperatureLevel"`
	HumidityLevel    string `json:"humidityLevel"`
	Coordinate       struct {
		Radial  float64 `json:"radial"`
		Angular int     `json:"angular"`
	} `json:"coordinate"`
}

// GetAirComfort returns the AirComfort for the active Home
func (c *APIClient) GetAirComfort(ctx context.Context) (airComfort AirComfort, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/airComfort", nil, &airComfort)
	}
	return airComfort, err
}
