package testutil_test

import (
	"github.com/clambin/tado"
	"github.com/clambin/tado/testutil"
	"reflect"
	"testing"
)

func TestMakeMobileDevice(t *testing.T) {
	tests := []struct {
		name    string
		options []testutil.MobileDeviceOption
		want    tado.MobileDevice
	}{
		{
			name:    "base",
			options: nil,
			want: tado.MobileDevice{
				ID:   1,
				Name: "foo",
			},
		},
		{
			name:    "home",
			options: []testutil.MobileDeviceOption{testutil.Home(true)},
			want: tado.MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: tado.MobileDeviceSettings{GeoTrackingEnabled: true},
				Location: tado.MobileDeviceLocation{AtHome: true},
			},
		},
		{
			name:    "away",
			options: []testutil.MobileDeviceOption{testutil.Home(false)},
			want: tado.MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: tado.MobileDeviceSettings{GeoTrackingEnabled: true},
			},
		},
		{
			name:    "stale",
			options: []testutil.MobileDeviceOption{testutil.Stale()},
			want: tado.MobileDevice{
				ID:       1,
				Name:     "foo",
				Settings: tado.MobileDeviceSettings{GeoTrackingEnabled: true},
				Location: tado.MobileDeviceLocation{Stale: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testutil.MakeMobileDevice(1, "foo", tt.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeMobileDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}
