package tado

import (
	"context"
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
	Setting     ZonePowerSetting           `json:"setting"`
	Termination ZoneInfoOverlayTermination `json:"termination"`
}

// ZonePowerSetting contains the zone's overlay settings
type ZonePowerSetting struct {
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
	return callAPI[ZoneInfo](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/state", nil)
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
	return callAPI[ZoneCapabilities](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/capabilities", nil)
}

// GetZoneEarlyStart checks if "early start" is enabled for the specified zone
func (c *APIClient) GetZoneEarlyStart(ctx context.Context, zoneID int) (earlyStart bool, err error) {
	type result struct {
		Enabled bool
	}
	response, err := callAPI[result](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/earlyStart", nil)
	if err == nil {
		earlyStart = response.Enabled
	}
	return
}

// SetZoneEarlyStart enabled or disables earlyStart for the specified zone
func (c *APIClient) SetZoneEarlyStart(ctx context.Context, zoneID int, earlyAccess bool) error {
	input := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: earlyAccess}

	_, err := callAPI[struct{}](ctx, c, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/earlyStart", input)
	return err
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
	Type         string            `json:"type"`
	AutoAdjust   bool              `json:"autoAdjust"`
	ComfortLevel AutoAdjustMode    `json:"comfortLevel"`
	Setting      *ZonePowerSetting `json:"setting"`
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
	return callAPI[ZoneAwayConfiguration](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/awayConfiguration", nil)
}

// SetZoneAutoConfiguration sets the ZoneAwayConfiguration for the specified zone
func (c *APIClient) SetZoneAutoConfiguration(ctx context.Context, zoneID int, configuration ZoneAwayConfiguration) (err error) {
	if configuration.AutoAdjust &&
		configuration.ComfortLevel != Eco &&
		configuration.ComfortLevel != Balance &&
		configuration.ComfortLevel != Comfort {
		return fmt.Errorf("invalid ComfortLevel: %d", configuration.ComfortLevel)
	}
	_, err = callAPI[struct{}](ctx, c, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/awayConfiguration", configuration)
	return
}
