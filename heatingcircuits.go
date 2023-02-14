package tado

import (
	"context"
	"net/http"
)

type HeatingCircuit struct {
	DriverSerialNo      string `json:"driverSerialNo"`
	DriverShortSerialNo string `json:"driverShortSerialNo"`
	Number              int    `json:"number"`
}

func (c *APIClient) GetHeatingCircuits(ctx context.Context) (output []HeatingCircuit, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/heatingCircuits", nil, &output)
	}
	return
}
