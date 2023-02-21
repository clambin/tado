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
	Preparation        interface{}        `json:"preparation"`
	Setting            ZonePowerSetting   `json:"setting"`
	OverlayType        string             `json:"overlayType"`
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
	ActivityDataPoints ZoneInfoActivityDataPoints `json:"activityDataPoints"`
	SensorDataPoints   ZoneInfoSensorDataPoints   `json:"sensorDataPoints"`
}

// ZoneInfoOpenWindow contains info on an open window. Only set if a window is open
type ZoneInfoOpenWindow struct {
	DetectedTime           time.Time `json:"detectedTime"`
	DurationInSeconds      int       `json:"durationInSeconds"`
	Expiry                 time.Time `json:"expiry"`
	RemainingTimeInSeconds int       `json:"remainingTimeInSeconds"`
}

// ZoneInfoOverlay contains the zone's manual settings.
//
// Tado supports three types of overlays: permanent ones (no expiry), timer-based ones (with a fixed time) and overlays
// that expire at the next block change on the timeTable. Use GetMode() to help determine the type of overlay.
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

// ZoneInfoOverlayTermination contains the termination parameters for the zone's overlay. Timers will only be populated for non-permanent modes.
type ZoneInfoOverlayTermination struct {
	Type                   string    `json:"type"`
	TypeSkillBasedApp      string    `json:"typeSkillBasedApp"`
	DurationInSeconds      int       `json:"durationInSeconds"`
	Expiry                 time.Time `json:"expiry"`
	RemainingTimeInSeconds int       `json:"remainingTimeInSeconds"`
	ProjectedExpiry        time.Time `json:"projectedExpiry"`
}

// ZoneInfoActivityDataPoints contains the zone's heating info
type ZoneInfoActivityDataPoints struct {
	HeatingPower Percentage `json:"heatingPower"`
}

// ZoneInfoSensorDataPoints contains the zone's current temperature & humidity
type ZoneInfoSensorDataPoints struct {
	InsideTemperature Temperature `json:"insideTemperature"`
	Humidity          Percentage  `json:"humidity"`
}

type OverlayTerminationMode int

const (
	UnknownOverlay OverlayTerminationMode = iota
	NoOverlay
	PermanentOverlay
	TimerOverlay
	NextBlockOverlay
)

// GetMode determines the type of overlay, i.e. permanent, timer-based or expiring at the next block change.
func (z ZoneInfoOverlay) GetMode() OverlayTerminationMode {
	if z.Type == "" {
		return NoOverlay
	}
	switch z.Termination.Type {
	case "MANUAL":
		return PermanentOverlay
	case "TIMER":
		switch z.Termination.TypeSkillBasedApp {
		case "TIMER":
			return TimerOverlay
		case "NEXT_TIME_BLOCK":
			return NextBlockOverlay
		}
	}
	return UnknownOverlay
}

// GetZoneInfo gets the info for the specified ZoneID
func (c *APIClient) GetZoneInfo(ctx context.Context, zoneID int) (tadoZoneInfo ZoneInfo, err error) {
	return callAPI[ZoneInfo](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/state", nil)
}
