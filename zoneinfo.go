package tado

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// ZoneCapabilities returns the "capabilities" of a Tado zone
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

// GetZoneCapabilities gets the capabilities for the specified zone
func (c *APIClient) GetZoneCapabilities(ctx context.Context, zoneID int) (tadoZoneCapabilities ZoneCapabilities, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/capabilities", nil, &tadoZoneCapabilities)
	}
	return
}

// GetZoneEarlyStart checks if "early start" is enabled for the specified zone
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

// SetZoneEarlyStart enabled or disables earlyStart for the specified zone
func (c *APIClient) SetZoneEarlyStart(ctx context.Context, zoneID int, earlyAccess bool) (err error) {
	if err = c.initialize(ctx); err == nil {
		input := struct {
			Enabled bool `json:"enabled"`
		}{Enabled: earlyAccess}

		var buf bytes.Buffer
		if err = json.NewEncoder(&buf).Encode(input); err == nil {
			err = c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/earlyStart", &buf, nil)
		}
	}
	return
}

// ZoneAwayConfiguration determines how a Zone will be heated when all users are away and the home is in "away" mode.
// If AdjustType is true, the zone's heating will be switched off. When the heating is switched back on is determined by ComfortLevel (Eco, Balance, Comfort).
// If AdjustType is false, the zone will be heated as per the Settings field.  E.g. using the following ZoneAwayConfigutation will heat the room at 16ºC:
//
//	{
//	 "type": "HEATING",
//	 "autoAdjust": false,
//	 "setting": {
//	   "type": "HEATING",
//	   "power": "ON",
//	   "temperature": {
//	     "celsius": 16,
//	   }
//	 }
//	}
type ZoneAwayConfiguration struct {
	Type         string             `json:"type"`
	AutoAdjust   bool               `json:"autoAdjust"`
	ComfortLevel AutoAdjustMode     `json:"comfortLevel"`
	Setting      *ZonePowerSettings `json:"setting"`
}

// ZonePowerSettings specifies how a zone should be heated when all users are away, and the zone is not in autoAdjust mode.
type ZonePowerSettings struct {
	Type        string      `json:"type"`
	Power       string      `json:"power"`
	Temperature Temperature `json:"temperature"`
}

// AutoAdjustMode determines how the heating should be switched back on when one or more users return home.
type AutoAdjustMode int

const (
	// Eco mode
	Eco AutoAdjustMode = 0
	// Balance mode
	Balance AutoAdjustMode = 50
	// Comfort mode
	Comfort AutoAdjustMode = 100
)

// GetZoneAutoConfiguration returns the ZoneAwayConfiguration for the specified zone
func (c *APIClient) GetZoneAutoConfiguration(ctx context.Context, zoneID int) (configuration ZoneAwayConfiguration, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/awayConfiguration", nil, &configuration)
	}
	return
}

// SetZoneAutoConfiguration sets the ZoneAwayConfiguration for the specified zone
func (c *APIClient) SetZoneAutoConfiguration(ctx context.Context, zoneID int, configuration ZoneAwayConfiguration) (err error) {
	if configuration.AutoAdjust &&
		configuration.ComfortLevel != Eco &&
		configuration.ComfortLevel != Balance &&
		configuration.ComfortLevel != Comfort {
		return fmt.Errorf("invalid ComfortLevel: %d", configuration.ComfortLevel)
	}
	if err = c.initialize(ctx); err == nil {
		payload := &bytes.Buffer{}
		if err = json.NewEncoder(payload).Encode(configuration); err == nil {
			err = c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/awayConfiguration", payload, nil)
		}
	}
	return
}

// Schedule is the heating schedule for a zone, i.e. the target temperature for the zone between a given start and end time.
type Schedule struct {
	DayType             string            `json:"dayType"`
	Start               string            `json:"start"`
	End                 string            `json:"end"`
	GeolocationOverride bool              `json:"geolocationOverride"`
	ModeID              int               `json:"modeId"`
	Setting             ZonePowerSettings `json:"setting"`
}

// GetZoneSchedule returns all Schedule entries for a zone
func (c *APIClient) GetZoneSchedule(ctx context.Context, zoneID int) (schedules []Schedule, err error) {
	if err = c.initialize(ctx); err == nil {
		// TODO: 1 is the schedule nr?
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/1/blocks", nil, &schedules)
	}
	return
}

// GetZoneScheduleForDay returns all Schedule entries for a zone for a given day
func (c *APIClient) GetZoneScheduleForDay(ctx context.Context, zoneID int, day string) (schedules []Schedule, err error) {
	if err = c.initialize(ctx); err == nil {
		// TODO: 1 is the schedule nr?
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/1/blocks/"+day, nil, &schedules)
	}
	return
}

// SetZoneScheduleForDay sets the Schedule entries for a zone for a given day
func (c *APIClient) SetZoneScheduleForDay(ctx context.Context, zoneID int, day string, schedules []Schedule) (err error) {
	if err = c.initialize(ctx); err == nil {
		var buf bytes.Buffer
		if err = json.NewEncoder(&buf).Encode(schedules); err == nil {
			// TODO: 1 is the schedule nr?
			err = c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/1/blocks/"+day, &buf, nil)
		}
	}
	return
}
