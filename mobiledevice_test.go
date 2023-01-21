package tado_test

import (
	"context"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetMobileDevices(t *testing.T) {
	info := []tado.MobileDevice{
		{
			ID:   1,
			Name: "foo",
			Settings: tado.MobileDeviceSettings{
				GeoTrackingEnabled: false,
			},
			Location: tado.MobileDeviceLocation{},
		},
		{
			ID:   2,
			Name: "bar",
			Settings: tado.MobileDeviceSettings{
				GeoTrackingEnabled: true,
			},
			Location: tado.MobileDeviceLocation{
				Stale:  false,
				AtHome: true,
			},
		},
	}

	c, s := makeTestServer(info)
	rcvd, err := c.GetMobileDevices(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info, rcvd)

	s.Close()
	_, err = c.GetMobileDevices(context.Background())
	assert.Error(t, err)
}

func TestMobileDevice_IsHome(t *testing.T) {
	tests := []struct {
		name   string
		device tado.MobileDevice
		state  tado.MobileDeviceLocationState
	}{
		{
			name: "home",
			device: tado.MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: tado.MobileDeviceSettings{GeoTrackingEnabled: true},
				Location: tado.MobileDeviceLocation{AtHome: true},
			},
			state: tado.DeviceHome,
		},
		{
			name: "home",
			device: tado.MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: tado.MobileDeviceSettings{GeoTrackingEnabled: true},
				Location: tado.MobileDeviceLocation{AtHome: false},
			},
			state: tado.DeviceAway,
		},
		{
			name: "no geotracking",
			device: tado.MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: tado.MobileDeviceSettings{GeoTrackingEnabled: false},
				Location: tado.MobileDeviceLocation{AtHome: false},
			},
			state: tado.DeviceUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.state, tt.device.IsHome())
		})
	}
}
