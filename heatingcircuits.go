package tado

import (
	"context"
	"net/http"
)

// HeatingCircuit contains details on a Tado heating circuit
type HeatingCircuit struct {
	DriverSerialNo      string `json:"driverSerialNo"`
	DriverShortSerialNo string `json:"driverShortSerialNo"`
	Number              int    `json:"number"`
}

// GetHeatingCircuits returns all registered heating circuits
func (c *APIClient) GetHeatingCircuits(ctx context.Context) (output []HeatingCircuit, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/heatingCircuits", nil, &output)
	}
	return
}
