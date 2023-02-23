package tado

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SetZoneOverlay sets an overlay (manual temperature setting) for the specified ZoneID
func (c *APIClient) SetZoneOverlay(ctx context.Context, zoneID int, temperature float64) (err error) {
	return c.SetZoneTemporaryOverlay(ctx, zoneID, temperature, 0)
}

// SetZoneTemporaryOverlay sets a temporary overlay (manual temperature setting) for the specified ZoneID for the specified amount of time.
// If duration is zero, it is equivalent to calling SetZoneOverlay().
func (c *APIClient) SetZoneTemporaryOverlay(ctx context.Context, zoneID int, temperature float64, duration time.Duration) (err error) {
	power := "ON"
	if temperature < 5 {
		temperature = 5
		power = "OFF"
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
		Setting: ZonePowerSetting{
			Type:        "HEATING",
			Power:       power,
			Temperature: Temperature{Celsius: temperature},
		},
		Termination: termination,
	}

	_, err = callAPI[ZoneInfoOverlay](ctx, c, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/overlay", request)
	return
}

// DeleteZoneOverlay deletes the overlay (manual temperature setting) for the specified ZoneID
func (c *APIClient) DeleteZoneOverlay(ctx context.Context, zoneID int) error {
	_, err := callAPI[struct{}](ctx, c, http.MethodDelete, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/overlay", nil)
	return err
}

// DefaultOverlay defines what happens when a user sets a manual temperature setting ("overlay") on the TRV.
// Type can be "MANUAL", "TADO_MODE" or "TIMER". For "TIMER", the DurationInSeconds must be set.
type DefaultOverlay struct {
	TerminationCondition struct {
		Type              string `json:"type"`
		DurationInSeconds int    `json:"durationInSeconds"`
	} `json:"terminationCondition"`
}

func (o DefaultOverlay) isValid() (string, bool) {
	switch o.TerminationCondition.Type {
	case "TADO_MODE":
	case "MANUAL":
	case "TIMER":
		if o.TerminationCondition.DurationInSeconds == 0 {
			return "DurationInSeconds must be set for TIMER overlay", false
		}
	default:
		return "invalid type: " + o.TerminationCondition.Type, false
	}
	return "", true
}

// GetDefaultOverlay returns the default overlay for the specified zone, which defines what happens when a user sets a manual temperature setting ("overlay") on the TRV.
func (c *APIClient) GetDefaultOverlay(ctx context.Context, zoneID int) (DefaultOverlay, error) {
	return callAPI[DefaultOverlay](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/defaultOverlay", nil)
}

// SetDefaultOverlay sets the DefaultOverlay for the specified zone, which defines what happens when a user sets a manual temperature setting ("overlay") on the TRV.
func (c *APIClient) SetDefaultOverlay(ctx context.Context, zoneID int, mode DefaultOverlay) error {
	if reason, valid := mode.isValid(); !valid {
		return fmt.Errorf("invalid overlay: %s", reason)
	}
	_, err := callAPI[struct{}](ctx, c, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/defaultOverlay", mode)
	return err
}
