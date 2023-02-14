package tado

// ZoneState is the state of the zone, i.e. heating is off, controlled automatically, or controlled manually
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
func (zoneInfo ZoneInfo) GetState() ZoneState {
	var state ZoneState
	if zoneInfo.Overlay.Type == "MANUAL" && zoneInfo.Overlay.Setting.Type == "HEATING" {
		if zoneInfo.Overlay.Setting.Power != "ON" || zoneInfo.Overlay.Setting.Temperature.Celsius <= 5.0 {
			state = ZoneStateOff
		} else if zoneInfo.Overlay.Termination.Type != "MANUAL" {
			state = ZoneStateTemporaryManual
		} else {
			state = ZoneStateManual
		}
	} else if zoneInfo.Setting.Power != "ON" {
		state = ZoneStateOff
	} else {
		state = ZoneStateAuto
	}
	return state
}
