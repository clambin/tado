package testutil

import "github.com/clambin/tado"

// MakeMobileDevice returns a tado.MobileDevice that has been configured with the provided options.
func MakeMobileDevice(id int, name string, options ...MobileDeviceOption) tado.MobileDevice {
	device := tado.MobileDevice{
		ID:   id,
		Name: name,
	}
	for _, option := range options {
		option(&device)
	}
	return device
}

// MobileDeviceOption configures a tado.MobileDevice.
type MobileDeviceOption func(device *tado.MobileDevice)

// Home marks the device's location as geo-tracked and either home or away.
func Home(home bool) MobileDeviceOption {
	return func(d *tado.MobileDevice) {
		d.Settings.GeoTrackingEnabled = true
		d.Location.Stale = false
		d.Location.AtHome = home
	}
}

// Stale marks the device's location as geo-tracked and stale.
func Stale() MobileDeviceOption {
	return func(d *tado.MobileDevice) {
		d.Settings.GeoTrackingEnabled = true
		d.Location.Stale = true
	}
}
