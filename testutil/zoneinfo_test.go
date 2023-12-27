package testutil_test

import (
	"github.com/clambin/tado"
	"github.com/clambin/tado/testutil"
	"reflect"
	"testing"
)

func TestMakeZoneInfo(t *testing.T) {
	tests := []struct {
		name    string
		options []testutil.ZoneInfoOption
		want    tado.ZoneInfo
	}{
		{
			name:    "on - auto",
			options: []testutil.ZoneInfoOption{testutil.ZoneInfoTemperature(19.0, 20.0)},
			want: tado.ZoneInfo{
				TadoMode: "HOME",
				Setting: tado.ZonePowerSetting{
					Type:        "HEATING",
					Power:       "ON",
					Temperature: tado.Temperature{Celsius: 20.0},
				},
				SensorDataPoints: tado.ZoneInfoSensorDataPoints{InsideTemperature: tado.Temperature{Celsius: 19.0}},
			},
		},
		{
			name:    "off - auto",
			options: []testutil.ZoneInfoOption{testutil.ZoneInfoTemperature(19.0, 5.0)},
			want: tado.ZoneInfo{
				TadoMode: "HOME",
				Setting: tado.ZonePowerSetting{
					Type:        "HEATING",
					Power:       "OFF",
					Temperature: tado.Temperature{Celsius: 5.0},
				},
				SensorDataPoints: tado.ZoneInfoSensorDataPoints{InsideTemperature: tado.Temperature{Celsius: 19.0}},
			},
		},
		{
			name:    "on - permanent overlay",
			options: []testutil.ZoneInfoOption{testutil.ZoneInfoTemperature(19.0, 20.0), testutil.ZoneInfoPermanentOverlay()},
			want: tado.ZoneInfo{
				TadoMode: "HOME",
				Setting:  tado.ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: tado.Temperature{Celsius: 20.0}},
				Overlay: tado.ZoneInfoOverlay{
					Type: "MANUAL",
					Termination: tado.ZoneInfoOverlayTermination{
						Type:              "MANUAL",
						TypeSkillBasedApp: "MANUAL",
					},
				},
				SensorDataPoints: tado.ZoneInfoSensorDataPoints{InsideTemperature: tado.Temperature{Celsius: 19.0}},
			},
		},
		{
			name:    "on - timer overlay",
			options: []testutil.ZoneInfoOption{testutil.ZoneInfoTemperature(19.0, 20.0), testutil.ZoneInfoTimerOverlay()},
			want: tado.ZoneInfo{
				TadoMode: "HOME",
				Setting:  tado.ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: tado.Temperature{Celsius: 20.0}},
				Overlay: tado.ZoneInfoOverlay{
					Type: "MANUAL",
					Termination: tado.ZoneInfoOverlayTermination{
						Type:              "TIMER",
						TypeSkillBasedApp: "TIMER",
					},
				},
				SensorDataPoints: tado.ZoneInfoSensorDataPoints{InsideTemperature: tado.Temperature{Celsius: 19.0}},
			},
		},
		{
			name:    "on - next_time_block overlay",
			options: []testutil.ZoneInfoOption{testutil.ZoneInfoTemperature(19.0, 20.0), testutil.ZoneInfoNextTimeBlockOverlay()},
			want: tado.ZoneInfo{
				TadoMode: "HOME",
				Setting:  tado.ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: tado.Temperature{Celsius: 20.0}},
				Overlay: tado.ZoneInfoOverlay{
					Type: "MANUAL",
					Termination: tado.ZoneInfoOverlayTermination{
						Type:              "TIMER",
						TypeSkillBasedApp: "NEXT_TIME_BLOCK",
					},
				},
				SensorDataPoints: tado.ZoneInfoSensorDataPoints{InsideTemperature: tado.Temperature{Celsius: 19.0}},
			},
		},
		{
			name:    "away",
			options: []testutil.ZoneInfoOption{testutil.ZoneInfoTadoMode(false)},
			want: tado.ZoneInfo{
				TadoMode: "AWAY",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := testutil.MakeZoneInfo(tt.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeZoneInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
