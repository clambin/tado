package tado

import (
	"context"
	"net/http"
	"time"
)

type AirComfort struct {
	Freshness struct {
		Value          string    `json:"value"`
		LastOpenWindow time.Time `json:"lastOpenWindow"`
	} `json:"freshness"`
	Comfort []ZoneAirComfort `json:"comfort"`
}

type ZoneAirComfort struct {
	RoomId           int    `json:"roomId"`
	TemperatureLevel string `json:"temperatureLevel"`
	HumidityLevel    string `json:"humidityLevel"`
	Coordinate       struct {
		Radial  float64 `json:"radial"`
		Angular int     `json:"angular"`
	} `json:"coordinate"`
}

func (c *APIClient) GetAirComfort(ctx context.Context) (airComfort AirComfort, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/airComfort", nil, &airComfort)
	}
	return airComfort, err
}
