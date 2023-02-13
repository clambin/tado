package tado

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// SetZoneOverlay sets an overlay (manual temperature setting) for the specified ZoneID
func (c *APIClient) SetZoneOverlay(ctx context.Context, zoneID int, temperature float64) (err error) {
	return c.SetZoneTemporaryOverlay(ctx, zoneID, temperature, 0)
}

// SetZoneTemporaryOverlay sets a temporary overlay (manual temperature setting) for the specified ZoneID for the specified amount of time.
// If duration is zero, it is equivalent to calling SetZoneOverlay()
func (c *APIClient) SetZoneTemporaryOverlay(ctx context.Context, zoneID int, temperature float64, duration time.Duration) (err error) {
	power := "ON"
	if temperature < 5 {
		temperature = 5
		power = "OFF"
	}

	if err = c.initialize(ctx); err != nil {
		return
	}

	var termination ZoneInfoOverlayTermination
	if duration > 0 {
		termination.Type = "TIMER"
		termination.DurationInSeconds = int(duration.Seconds())
	} else {
		termination.Type = "MANUAL"
	}

	request := ZoneInfoOverlay{
		Type: "MANUAL",
		Setting: ZoneInfoOverlaySetting{
			Type:  "HEATING",
			Power: power,
			Temperature: Temperature{
				Celsius: temperature,
			},
		},
		Termination: termination,
	}

	var payload bytes.Buffer
	if err = json.NewEncoder(&payload).Encode(request); err == nil {
		err = c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/overlay", &payload, nil)
	}

	return
}

// DeleteZoneOverlay deletes the overlay (manual temperature setting) for the specified ZoneID
func (c *APIClient) DeleteZoneOverlay(ctx context.Context, zoneID int) (err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodDelete, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/overlay", nil, nil)
	}
	return
}
