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
	Type         string                         `json:"type"`
	AutoAdjust   bool                           `json:"autoAdjust"`
	ComfortLevel AutoAdjustMode                 `json:"comfortLevel"`
	Setting      *ZoneAwayConfigurationSettings `json:"setting"`
}

// ZoneAwayConfigurationSettings specifies how a zone should be heated when all users are away, and the zone is not in autoAdjust mode.
type ZoneAwayConfigurationSettings struct {
	Type        string      `json:"type"`
	Power       string      `json:"power"`
	Temperature Temperature `json:"temperature"`
}

// AutoAdjustMode determines how the heating should be switched back on when one or more users return home.
type AutoAdjustMode int

const (
	Eco     AutoAdjustMode = 0
	Balance AutoAdjustMode = 50
	Comfort AutoAdjustMode = 100
)

func (c *APIClient) GetZoneAutoConfiguration(ctx context.Context, zoneID int) (configuration ZoneAwayConfiguration, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/awayConfiguration", nil, &configuration)
	}
	return
}

func (c *APIClient) SetZoneAutoConfiguration(ctx context.Context, zoneID int, configuration ZoneAwayConfiguration) (err error) {
	if err = c.initialize(ctx); err == nil {
		payload := &bytes.Buffer{}
		if err = json.NewEncoder(payload).Encode(configuration); err == nil {
			err = c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/awayConfiguration", payload, nil)
		}
	}
	return
}
