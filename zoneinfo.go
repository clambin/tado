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
type ZoneInfo struct {
	TadoMode            string `json:"tadoMode"`
	GeolocationOverride bool   `json:"geolocationOverride"`
	// TODO
	GeolocationOverrideDisableTime interface{} `json:"geolocationOverrideDisableTime"`
	// TODO
	Preparation        interface{}                `json:"preparation"`
	Setting            ZoneInfoSetting            `json:"setting"`
	ActivityDataPoints ZoneInfoActivityDataPoints `json:"activityDataPoints"`
	SensorDataPoints   ZoneInfoSensorDataPoints   `json:"sensorDataPoints"`
	// TODO
	OverlayType        interface{}        `json:"overlayType"`
	Overlay            ZoneInfoOverlay    `json:"overlay,omitempty"`
	OpenWindow         ZoneInfoOpenWindow `json:"openwindow,omitempty"`
	NextScheduleChange struct {
		Start   time.Time `json:"start"`
		Setting struct {
			Type        string      `json:"type"`
			Power       string      `json:"power"`
			Temperature Temperature `json:"temperature"`
		} `json:"setting"`
	} `json:"nextScheduleChange"`
	NextTimeBlock struct {
		Start time.Time `json:"start"`
	} `json:"nextTimeBlock"`
	Link struct {
		State string `json:"state"`
	} `json:"link"`
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
func (c *APIClient) GetZoneInfo(ctx context.Context, zoneID int) (tadoZoneInfo ZoneInfo, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/state", nil, &tadoZoneInfo)
	}
	return
}

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

func (s ZoneState) String() string {
	names := map[ZoneState]string{
		ZoneStateUnknown:         "unknown",
		ZoneStateOff:             "off",
		ZoneStateAuto:            "auto",
		ZoneStateTemporaryManual: "manual (temp)",
		ZoneStateManual:          "manual",
	}
	name, ok := names[s]
	if !ok {
		name = "(invalid)"
	}
	return name
}

// GetState returns the state of the zone
func (zoneInfo ZoneInfo) GetState() ZoneState {
	var state ZoneState
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
	return state
}
