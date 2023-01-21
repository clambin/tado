package tado

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// ZoneInfo contains the response to /api/v2/homes/<HomeID>/zones/<zoneID>/state
//
// This structure provides the following key information:
//
//	Setting.Power:                              power state of the specified zone (0-1)
//	Temperature.Celsius:                        target temperature for the zone, in degrees Celsius
//	OpenWindow.DurationInSeconds:               how long an open window has been detected in seconds
//	ActivityDataPoints.HeatingPower.Percentage: heating power for the zone (0-100%)
//	SensorDataPoints.Temperature.Celsius:       current temperature, in degrees Celsius
//	SensorDataPoints.Humidity.Percentage:       humidity (0-100%)
type ZoneInfo struct {
	Setting            ZoneInfoSetting            `json:"setting"`
	ActivityDataPoints ZoneInfoActivityDataPoints `json:"activityDataPoints"`
	SensorDataPoints   ZoneInfoSensorDataPoints   `json:"sensorDataPoints"`
	OpenWindow         ZoneInfoOpenWindow         `json:"openwindow,omitempty"`
	Overlay            ZoneInfoOverlay            `json:"overlay,omitempty"`
}

// ZoneInfoSetting contains the zone's current power & target temperature
type ZoneInfoSetting struct {
	Power       string      `json:"power"`
	Temperature Temperature `json:"temperature"`
}

// ZoneInfoActivityDataPoints contains the zone's heating info
type ZoneInfoActivityDataPoints struct {
	HeatingPower Percentage `json:"heatingPower"`
}

// ZoneInfoSensorDataPoints contains the zone's current temperature & humidity
type ZoneInfoSensorDataPoints struct {
	Temperature Temperature `json:"insideTemperature"`
	Humidity    Percentage  `json:"humidity"`
}

// ZoneInfoOpenWindow contains info on an open window. Only set if a window is open
type ZoneInfoOpenWindow struct {
	DetectedTime           time.Time `json:"detectedTime"`
	DurationInSeconds      int       `json:"durationInSeconds"`
	Expiry                 time.Time `json:"expiry"`
	RemainingTimeInSeconds int       `json:"remainingTimeInSeconds"`
}

// ZoneInfoOverlay contains the zone's manual settings
type ZoneInfoOverlay struct {
	Type        string                     `json:"type"`
	Setting     ZoneInfoOverlaySetting     `json:"setting"`
	Termination ZoneInfoOverlayTermination `json:"termination"`
}

// ZoneInfoOverlaySetting contains the zone's overlay settings
type ZoneInfoOverlaySetting struct {
	Type        string      `json:"type"`
	Power       string      `json:"power"`
	Temperature Temperature `json:"temperature"`
}

// ZoneInfoOverlayTermination contains the termination parameters for the zone's overlay
type ZoneInfoOverlayTermination struct {
	Type              string `json:"type"`
	RemainingTime     int    `json:"remainingTimeInSeconds,omitempty"`
	DurationInSeconds int    `json:"durationInSeconds,omitempty"`
	// not specified:
	//  "typeSkillBasedApp":"NEXT_TIME_BLOCK",
	//  "expiry":"2021-02-05T23:00:00Z",
	//  "projectedExpiry":"2021-02-05T23:00:00Z"
}

// GetZoneInfo gets the info for the specified ZoneID
func (client *APIClient) GetZoneInfo(ctx context.Context, zoneID int) (tadoZoneInfo ZoneInfo, err error) {
	if err = client.initialize(ctx); err == nil {
		err = client.call(ctx, http.MethodGet, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/state"), nil, &tadoZoneInfo)
	}
	return
}

// SetZoneOverlay sets an overlay (manual temperature setting) for the specified ZoneID
func (client *APIClient) SetZoneOverlay(ctx context.Context, zoneID int, temperature float64) (err error) {
	if err = client.initialize(ctx); err != nil {
		return
	}

	if temperature < 5 {
		temperature = 5
	}
	request := ZoneInfoOverlay{
		Type: "MANUAL",
		Setting: ZoneInfoOverlaySetting{
			Type:  "HEATING",
			Power: "ON",
			Temperature: Temperature{
				Celsius: temperature,
			},
		},
		Termination: ZoneInfoOverlayTermination{
			Type: "MANUAL",
		},
	}
	var payload bytes.Buffer
	if err = json.NewEncoder(&payload).Encode(request); err == nil {
		err = client.call(ctx, http.MethodPut, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/overlay"), &payload, nil)
	}
	return

}

// SetZoneOverlayWithDuration sets an overlay (manual temperature setting) for the specified ZoneID for a specified amount of time.
// If duration is zero, it calls SetZoneOverlay.
func (client *APIClient) SetZoneOverlayWithDuration(ctx context.Context, zoneID int, temperature float64, duration time.Duration) (err error) {
	if duration == 0 {
		return client.SetZoneOverlay(ctx, zoneID, temperature)
	}

	if temperature < 5 {
		temperature = 5
	}

	if err = client.initialize(ctx); err != nil {
		return
	}

	request := ZoneInfoOverlay{
		Type: "MANUAL",
		Setting: ZoneInfoOverlaySetting{
			Type:  "HEATING",
			Power: "ON",
			Temperature: Temperature{
				Celsius: temperature,
			},
		},
		Termination: ZoneInfoOverlayTermination{
			Type:              "TIMER",
			DurationInSeconds: int(duration.Seconds()),
		},
	}
	var payload bytes.Buffer
	if err = json.NewEncoder(&payload).Encode(request); err == nil {
		err = client.call(ctx, http.MethodPut, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/overlay"), &payload, nil)
	}

	return
}

// DeleteZoneOverlay deletes the overlay (manual temperature setting) for the specified ZoneID
func (client *APIClient) DeleteZoneOverlay(ctx context.Context, zoneID int) (err error) {
	if err = client.initialize(ctx); err == nil {
		err = client.call(ctx, http.MethodDelete, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/overlay"), nil, nil)
	}

	return
}

// ZoneState is the state of the zone, i.e. heating is off, controlled automatically, or controlled manually
type ZoneState int

const (
	// ZoneStateUnknown indicates the zone's state is not initialized yet
	ZoneStateUnknown ZoneState = iota
	// ZoneStateOff indicates the zone's heating is switched off
	ZoneStateOff
	// ZoneStateAuto indicates the zone's heating is controlled manually (e.g. as per schedule)
	ZoneStateAuto
	// ZoneStateTemporaryManual indicates the zone's target temperature is set manually, for a period of time
	ZoneStateTemporaryManual
	// ZoneStateManual indicates the zone's target temperature is set manually
	ZoneStateManual
)

// GetState returns the state of the zone
func (zoneInfo ZoneInfo) GetState() (state ZoneState) {
	state = ZoneStateUnknown
	if zoneInfo.Overlay.Type == "MANUAL" && zoneInfo.Overlay.Setting.Type == "HEATING" {
		if zoneInfo.Overlay.Setting.Power != "ON" || zoneInfo.Overlay.Setting.Temperature.Celsius <= 5.0 {
			state = ZoneStateOff
		} else if zoneInfo.Overlay.Termination.Type != "MANUAL" {
			state = ZoneStateTemporaryManual
		} else {
			state = ZoneStateManual
		}
	} else if zoneInfo.Setting.Power != "ON" {
		state = ZoneStateOff
	} else {
		state = ZoneStateAuto
	}
	return
}
