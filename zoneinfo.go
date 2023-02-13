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

type ZoneCapabilities struct {
	Temperatures struct {
		Celsius struct {
			Max  int     `json:"max"`
			Min  int     `json:"min"`
			Step float64 `json:"step"`
		} `json:"celsius"`
		Fahrenheit struct {
			Max  int     `json:"max"`
			Min  int     `json:"min"`
			Step float64 `json:"step"`
		} `json:"fahrenheit"`
	} `json:"temperatures"`
	Type string `json:"type"`
}

// GetZoneCapabilities gets the capabilities for the specified ZoneID
func (c *APIClient) GetZoneCapabilities(ctx context.Context, zoneID int) (tadoZoneCapabilities ZoneCapabilities, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/capabilities", nil, &tadoZoneCapabilities)
	}
	return
}

func (c *APIClient) GetZoneEarlyStart(ctx context.Context, zoneID int) (earlyStart bool, err error) {
	if err = c.initialize(ctx); err == nil {
		var result struct {
			Enabled bool
		}
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/earlyStart", nil, &result)
		earlyStart = result.Enabled

	}
	return
}

func (c *APIClient) SetZoneEarlyStart(ctx context.Context, zoneID int, earlyAccess bool) error {
	if err := c.initialize(ctx); err != nil {
		return err
	}
	input := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: earlyAccess}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(input)
	if err != nil {
		return err
	}
	return c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/earlyStart", &buf, nil)
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
