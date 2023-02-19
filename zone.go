package tado

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Zone contains the configuration of a given zone
type Zone struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	DateCreated       time.Time `json:"dateCreated"`
	DeviceTypes       []string  `json:"deviceTypes"`
	Devices           []Device  `json:"devices"`
	ReportAvailable   bool      `json:"reportAvailable"`
	ShowScheduleSetup bool      `json:"showScheduleSetup"`
	SupportsDazzle    bool      `json:"supportsDazzle"`
	DazzleEnabled     bool      `json:"dazzleEnabled"`
	DazzleMode        struct {
		Supported bool `json:"supported"`
		Enabled   bool `json:"enabled"`
	} `json:"dazzleMode"`
	OpenWindowDetection struct {
		Supported        bool `json:"supported"`
		Enabled          bool `json:"enabled"`
		TimeoutInSeconds int  `json:"timeoutInSeconds"`
	} `json:"openWindowDetection"`
}

// Device contains attributes of a Tado device
type Device struct {
	DeviceType       string `json:"deviceType"`
	SerialNo         string `json:"serialNo"`
	ShortSerialNo    string `json:"shortSerialNo"`
	CurrentFwVersion string `json:"currentFwVersion"`
	ConnectionState  State  `json:"connectionState"`
	Characteristics  struct {
		Capabilities []string `json:"capabilities"`
	} `json:"characteristics"`
	InPairingMode          bool     `json:"inPairingMode,omitempty"`
	BatteryState           string   `json:"batteryState"`
	Duties                 []string `json:"duties"`
	MountingState          Value    `json:"mountingState,omitempty"`
	MountingStateWithError string   `json:"mountingStateWithError,omitempty"`
	ChildLockEnabled       bool     `json:"childLockEnabled,omitempty"`
}

// State contains the connection state of a Tado device
type State struct {
	Value     bool      `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// GetZones retrieves the different Zones configured for the user's Home ID
func (c *APIClient) GetZones(ctx context.Context) (zones Zones, err error) {
	return callAPI[Zones](ctx, c, http.MethodGet, "myTado", "/zones", nil)
}

// Zones contains a list of Zone records
type Zones []Zone

// GetZone retrieves the Zone from a list of Zones by ID. bool is false if the zone could not be found.
func (z Zones) GetZone(id int) (Zone, bool) {
	for _, zone := range z {
		if zone.ID == id {
			return zone, true
		}
	}
	return Zone{}, false
}

// GetZoneByName retrieves the Zone from a list of Zones by Name. ok is false if the zone could not be found
func (z Zones) GetZoneByName(name string) (Zone, bool) {
	for _, zone := range z {
		if zone.Name == name {
			return zone, true
		}
	}
	return Zone{}, false
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

// ZoneMeasuringDevice contains configuration parameters of a measuring device at a given zone
type ZoneMeasuringDevice struct {
	BatteryState    string `json:"batteryState"`
	Characteristics struct {
		Capabilities []string `json:"capabilities"`
	} `json:"characteristics"`
	ConnectionState  State
	CurrentFwVersion string `json:"currentFwVersion"`
	DeviceType       string `json:"deviceType"`
	SerialNo         string `json:"serialNo"`
	ShortSerialNo    string `json:"shortSerialNo"`
}

// GetZoneMeasuringDevice returns information on the measuring device at the specified zone
func (c *APIClient) GetZoneMeasuringDevice(ctx context.Context, zoneID int) (measuringDevice ZoneMeasuringDevice, err error) {
	return callAPI[ZoneMeasuringDevice](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/measuringDevice", nil)
}
