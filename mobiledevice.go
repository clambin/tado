package tado

import (
	"context"
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
func (c *APIClient) GetMobileDevices(ctx context.Context) (tadoMobileDevices []MobileDevice, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/mobileDevices", nil, &tadoMobileDevices)
	}
	return
}

// MobileDeviceLocationState is the state of the user's device (mobile phone), i.e. home or away
type MobileDeviceLocationState int

const (
	// DeviceUnknown means the device state is not initialized yet
	DeviceUnknown MobileDeviceLocationState = iota
	// DeviceHome means the user's device is home
	DeviceHome
	// DeviceAway means the user's device is not home
	DeviceAway
)

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
