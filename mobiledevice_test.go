package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetMobileDevices(t *testing.T) {
	info := []MobileDevice{
		{
			ID:   1,
			Name: "foo",
			Settings: MobileDeviceSettings{
				GeoTrackingEnabled: false,
			},
			Location: MobileDeviceLocation{},
		},
		{
			ID:   2,
			Name: "bar",
			Settings: MobileDeviceSettings{
				GeoTrackingEnabled: true,
			},
			Location: MobileDeviceLocation{
				Stale:  false,
				AtHome: true,
			},
		},
	}

	c, s := makeTestServer(info, nil)
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
		device MobileDevice
		state  MobileDeviceLocationState
	}{
		{
			name: "home",
			device: MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: MobileDeviceSettings{GeoTrackingEnabled: true},
				Location: MobileDeviceLocation{AtHome: true},
			},
			state: DeviceHome,
		},
		{
			name: "home",
			device: MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: MobileDeviceSettings{GeoTrackingEnabled: true},
				Location: MobileDeviceLocation{AtHome: false},
			},
			state: DeviceAway,
		},
		{
			name: "no geotracking",
			device: MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: MobileDeviceSettings{GeoTrackingEnabled: false},
				Location: MobileDeviceLocation{AtHome: false},
			},
			state: DeviceUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.state, tt.device.IsHome())
		})
	}
}
