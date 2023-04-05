package testutil

import "github.com/clambin/tado"

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

type MobileDeviceOption func(device *tado.MobileDevice)

func Home(home bool) MobileDeviceOption {
	return func(d *tado.MobileDevice) {
		d.Settings.GeoTrackingEnabled = true
		d.Location.Stale = false
		d.Location.AtHome = home
	}
}

func Stale() MobileDeviceOption {
	return func(d *tado.MobileDevice) {
		d.Settings.GeoTrackingEnabled = true
		d.Location.Stale = true
	}
}
