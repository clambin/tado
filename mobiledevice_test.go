package tado_test

import (
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMobileDevice_IsHome(t *testing.T) {
	device := tado.MobileDevice{
		ID:   1,
		Name: "foo",
		Settings: tado.MobileDeviceSettings{
			GeoTrackingEnabled: false,
		},
		Location: tado.MobileDeviceLocation{
			AtHome: false,
		},
	}

	assert.Equal(t, tado.DeviceUnknown, device.IsHome())

	device.Settings.GeoTrackingEnabled = true
	assert.Equal(t, tado.DeviceAway, device.IsHome())

	device.Location.AtHome = true
	assert.Equal(t, tado.DeviceHome, device.IsHome())
}
