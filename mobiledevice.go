package tado

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MobileDevice contains the response to /api/v2/homes/<HomeID>/mobileDevices
type MobileDevice struct {
	ID       int                  `json:"id"`
	Name     string               `json:"name"`
	Settings MobileDeviceSettings `json:"settings"`
	Location MobileDeviceLocation `json:"location"`
}

// MobileDeviceSettings is a sub-structure of MobileDevice
type MobileDeviceSettings struct {
	GeoTrackingEnabled bool `json:"geoTrackingEnabled"`
}

// MobileDeviceLocation is a sub-structure of MobileDevice
type MobileDeviceLocation struct {
	Stale  bool `json:"stale"`
	AtHome bool `json:"atHome"`
}

// GetMobileDevices retrieves the status of all registered mobile devices.
func (client *APIClient) GetMobileDevices(ctx context.Context) ([]MobileDevice, error) {
	var (
		err               error
		tadoMobileDevices []MobileDevice
		body              []byte
	)
	if err = client.initialize(ctx); err == nil {
		if body, err = client.call(ctx, http.MethodGet, client.apiV2URL("/mobileDevices"), ""); err == nil {
			err = json.Unmarshal(body, &tadoMobileDevices)
		}
	}

	return tadoMobileDevices, err
}

// String serializes a MobileDevice into a string. Used for logging.
func (mobileDevice *MobileDevice) String() string {
	return fmt.Sprintf("name=%s, geotrack=%v, stale=%v, athome=%v",
		mobileDevice.Name,
		mobileDevice.Settings.GeoTrackingEnabled,
		mobileDevice.Location.Stale,
		mobileDevice.Location.AtHome,
	)
}

const (
	// DeviceUnknown means the device state is not initialized yet
	DeviceUnknown = iota
	// DeviceHome means the user's device is home
	DeviceHome
	// DeviceAway means the user's device is not home
	DeviceAway
)

// MobileDeviceLocationState is the state of the user's device (mobile phone), i.e. home or away
type MobileDeviceLocationState int

// IsHome returns the location of the MobileDevice
func (mobileDevice *MobileDevice) IsHome() (state MobileDeviceLocationState) {
	state = DeviceUnknown
	if mobileDevice.Settings.GeoTrackingEnabled {
		if mobileDevice.Location.AtHome {
			state = DeviceHome
		} else {
			state = DeviceAway
		}
	}
	return
}
