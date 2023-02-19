package tado

import (
	"context"
	"net/http"
	"time"
)

// Zone contains the response to /api/v2/homes/<HomeID>/zones
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
	DeviceType       string          `json:"deviceType"`
	SerialNo         string          `json:"serialNo"`
	ShortSerialNo    string          `json:"shortSerialNo"`
	CurrentFwVersion string          `json:"currentFwVersion"`
	ConnectionState  ConnectionState `json:"connectionState"`
	Characteristics  struct {
		Capabilities []string `json:"capabilities"`
	} `json:"characteristics"`
	InPairingMode bool   `json:"inPairingMode,omitempty"`
	BatteryState  string `json:"batteryState"`
	MountingState struct {
		Value     string    `json:"value"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"mountingState,omitempty"`
	MountingStateWithError string `json:"mountingStateWithError,omitempty"`
	ChildLockEnabled       bool   `json:"childLockEnabled,omitempty"`
}

// ConnectionState contains the connection state of a Tado device
type ConnectionState struct {
	Value     bool      `json:"value"`
	Timestamp time.Time `json:"timeStamp"`
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
