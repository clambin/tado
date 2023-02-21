package tado

// ZoneState is the state of the zone, i.e. heating is off, controlled automatically, or controlled manually
//
// Deprecated: will be removed in a future release
type ZoneState int

const (
	// ZoneStateUnknown indicates the zone's state is not initialized yet
	ZoneStateUnknown ZoneState = iota
	// ZoneStateOff indicates the zone's heating is switched off
	ZoneStateOff
	// ZoneStateAuto indicates the zone's heating is controlled manually (e.g. as per schedule)
	ZoneStateAuto
	// ZoneStateTemporaryManual indicates the zone's target temperature is set manually, for a period of time
	ZoneStateTemporaryManual
	// ZoneStateManual indicates the zone's target temperature is set manually
	ZoneStateManual
)

// String returns a string representation of a ZoneState
//
// Deprecated: will be removed in a future release
func (s ZoneState) String() string {
	names := map[ZoneState]string{
		ZoneStateUnknown:         "unknown",
		ZoneStateOff:             "off",
		ZoneStateAuto:            "auto",
		ZoneStateTemporaryManual: "manual (temp)",
		ZoneStateManual:          "manual",
	}
	name, ok := names[s]
	if !ok {
		name = "(invalid)"
	}
	return name
}

// GetState returns the state of the zone
//
// Deprecated: will be removed in a future release
func (z ZoneInfo) GetState() ZoneState {
	if z.Setting.Power == "OFF" {
		return ZoneStateOff
	}
	switch z.Overlay.GetMode() {
	case NoOverlay:
		return ZoneStateAuto
	case PermanentOverlay:
		return ZoneStateManual
	case TimerOverlay, NextBlockOverlay:
		return ZoneStateTemporaryManual
	}
	return ZoneStateUnknown
}
