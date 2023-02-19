package tado

import (
	"context"
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
