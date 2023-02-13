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
