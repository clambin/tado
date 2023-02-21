package tado

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZoneState_String(t *testing.T) {
	tests := []struct {
		name string
		s    ZoneState
		want string
	}{
		{name: "ZoneStateUnknown", s: ZoneStateUnknown, want: "unknown"},
		{name: "ZoneStateOff", s: ZoneStateOff, want: "off"},
		{name: "ZoneStateAuto", s: ZoneStateAuto, want: "auto"},
		{name: "ZoneStateTemporaryManual", s: ZoneStateTemporaryManual, want: "manual (temp)"},
		{name: "ZoneStateManual", s: ZoneStateManual, want: "manual"},
		{name: "invalid", s: ZoneState(-1), want: "(invalid)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.String(), "String()")
		})
	}
}

func TestZoneInfo_GetState(t *testing.T) {
	tests := []struct {
		name     string
		zoneInfo ZoneInfo
		state    ZoneState
	}{
		{
			name: "auto",
			zoneInfo: ZoneInfo{
				Setting: ZonePowerSetting{Power: "ON", Temperature: Temperature{Celsius: 17.0}},
			},
			state: ZoneStateAuto,
		},
		{
			name: "manual",
			zoneInfo: ZoneInfo{
				Setting: ZonePowerSetting{Power: "ON", Temperature: Temperature{Celsius: 22.0}},
				Overlay: ZoneInfoOverlay{
					Type:        "MANUAL",
					Termination: ZoneInfoOverlayTermination{Type: "MANUAL"},
				},
			},
			state: ZoneStateManual,
		},
		{
			name: "manual w/ termination",
			zoneInfo: ZoneInfo{
				Setting: ZonePowerSetting{Power: "ON", Temperature: Temperature{Celsius: 17.0}},
				Overlay: ZoneInfoOverlay{
					Type:        "MANUAL",
					Termination: ZoneInfoOverlayTermination{Type: "TIMER", TypeSkillBasedApp: "NEXT_TIME_BLOCK"},
				},
			},
			state: ZoneStateTemporaryManual,
		},
		{
			name: "off",
			zoneInfo: ZoneInfo{
				Setting: ZonePowerSetting{Power: "OFF"},
			},
			state: ZoneStateOff,
		},
		{
			name: "manual off",
			zoneInfo: ZoneInfo{
				Setting: ZonePowerSetting{Power: "OFF"},
				Overlay: ZoneInfoOverlay{
					Type:        "MANUAL",
					Termination: ZoneInfoOverlayTermination{Type: "TIMER", TypeSkillBasedApp: "TIMER"},
				},
			},
			state: ZoneStateOff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.state, tt.zoneInfo.GetState())
		})
	}
}
