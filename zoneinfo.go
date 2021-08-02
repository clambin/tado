package tado

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ZoneInfo contains the response to /api/v2/homes/<HomeID>/zones/<zoneID>/state
//
// This structure provides the following key information:
//   Setting.Power:                              power state of the specified zone (0-1)
//   Temperature.Celsius:                        target temperature for the zone, in degrees Celsius
//   OpenWindow.DurationInSeconds:               how long an open window has been detected in seconds
//   ActivityDataPoints.HeatingPower.Percentage: heating power for the zone (0-100%)
//   SensorDataPoints.Temperature.Celsius:       current temperature, in degrees Celsius
//   SensorDataPoints.Humidity.Percentage:       humidity (0-100%)
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
func (client *APIClient) GetZoneInfo(ctx context.Context, zoneID int) (ZoneInfo, error) {
	var (
		err          error
		body         []byte
		tadoZoneInfo ZoneInfo
	)
	if err = client.initialize(ctx); err == nil {
		if body, err = client.call(ctx, http.MethodGet, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/state"), ""); err == nil {
			err = json.Unmarshal(body, &tadoZoneInfo)
		}
	}
	return tadoZoneInfo, err
}

// SetZoneOverlay sets an overlay (manual temperature setting) for the specified ZoneID
func (client *APIClient) SetZoneOverlay(ctx context.Context, zoneID int, temperature float64) error {
	if temperature < 5 {
		temperature = 5
	}

	var (
		err     error
		payload []byte
	)

	if err = client.initialize(ctx); err == nil {
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

		payload, err = json.Marshal(request)

		_, err = client.call(ctx, http.MethodPut, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/overlay"), string(payload))
	}

	return err
}

// SetZoneOverlayWithDuration sets an overlay (manual temperature setting) for the specified ZoneID for a specified amount of time.
// If duration is zero, it calls SetZoneOverlay.
func (client *APIClient) SetZoneOverlayWithDuration(ctx context.Context, zoneID int, temperature float64, duration time.Duration) error {
	if duration == 0 {
		return client.SetZoneOverlay(ctx, zoneID, temperature)
	}

	if temperature < 5 {
		temperature = 5
	}

	var (
		err     error
		payload []byte
	)

	if err = client.initialize(ctx); err == nil {
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

		payload, err = json.Marshal(request)

		_, err = client.call(ctx, http.MethodPut, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/overlay"), string(payload))
	}

	return err
}

// DeleteZoneOverlay deletes the overlay (manual temperature setting) for the specified ZoneID
func (client *APIClient) DeleteZoneOverlay(ctx context.Context, zoneID int) error {
	var err error
	if err = client.initialize(ctx); err == nil {
		_, err = client.call(ctx, http.MethodDelete, client.apiV2URL("/zones/"+strconv.Itoa(zoneID)+"/overlay"), "")
	}

	return err
}

// String serializes a ZoneInfo into a string. Used for logging.
func (zoneInfo *ZoneInfo) String() string {
	return fmt.Sprintf("target=%.1fºC, temp=%.1fºC, humidity=%.1f%%, heating=%.1f%%, power=%s, openwindow=%ds, overlay={%s}",
		zoneInfo.Setting.Temperature.Celsius,
		zoneInfo.SensorDataPoints.Temperature.Celsius,
		zoneInfo.SensorDataPoints.Humidity.Percentage,
		zoneInfo.ActivityDataPoints.HeatingPower.Percentage,
		zoneInfo.Setting.Power,
		zoneInfo.OpenWindow.DurationInSeconds-zoneInfo.OpenWindow.RemainingTimeInSeconds,
		zoneInfo.Overlay.String(),
	)
}

// String serializes a ZoneInfoOverlay into a string. Used for logging.
func (overlay *ZoneInfoOverlay) String() string {
	return fmt.Sprintf(`type=%s, settings={%s}, termination={type="%s", remaining=%d}`,
		overlay.Type,
		overlay.Setting.String(),
		overlay.Termination.Type,
		overlay.Termination.RemainingTime,
	)
}

// String serializes a ZoneInfoOverlaySetting into a string. Used for logging.
func (setting *ZoneInfoOverlaySetting) String() string {
	return fmt.Sprintf("type=%s, power=%s, temp=%.1fºC",
		setting.Type,
		setting.Power,
		setting.Temperature.Celsius,
	)
}

const (
	ZoneStateUnknown = iota
	ZoneStateOff
	ZoneStateAuto
	ZoneStateTemporaryManual
	ZoneStateManual
)

type ZoneState int

// GetState returns the state of the zone, i.e.
func (zoneInfo *ZoneInfo) GetState() (state ZoneState) {
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
