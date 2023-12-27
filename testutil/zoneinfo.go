package testutil

import (
	"github.com/clambin/tado"
)

// MakeZoneInfo creates a tado.ZoneInfo. The provided options are used to configure the resulting zoneInfo.
func MakeZoneInfo(options ...ZoneInfoOption) tado.ZoneInfo {
	zoneInfo := tado.ZoneInfo{TadoMode: "HOME"}

	for _, option := range options {
		option(&zoneInfo)
	}

	return zoneInfo
}

// ZoneInfoOption configures a tado.ZoneInfo.
type ZoneInfoOption func(*tado.ZoneInfo)

// ZoneInfoTadoMode sets the TadoMode (home/away) in a tado.ZoneInfo.
func ZoneInfoTadoMode(home bool) ZoneInfoOption {
	return func(z *tado.ZoneInfo) {
		if home {
			z.TadoMode = "HOME"
		} else {
			z.TadoMode = "AWAY"
		}
	}
}

// ZoneInfoTemperature sets the measured & target temperatures in a tado.ZoneInfo.
func ZoneInfoTemperature(measured, target float64) ZoneInfoOption {
	return func(z *tado.ZoneInfo) {
		z.SensorDataPoints.InsideTemperature.Celsius = measured
		z.Setting.Type = "HEATING"
		if target <= 5.0 {
			z.Setting.Power = "OFF"
		} else {
			z.Setting.Power = "ON"
		}
		z.Setting.Temperature.Celsius = target
	}
}

func zoneInfoOverlay(overlayType, overlaySubType string) ZoneInfoOption {
	return func(z *tado.ZoneInfo) {
		z.Overlay.Type = "MANUAL"
		z.Overlay.Termination.Type = overlayType
		z.Overlay.Termination.TypeSkillBasedApp = overlaySubType
	}
}

// ZoneInfoPermanentOverlay adds a permanent overlay to the tado.ZoneInfo.
func ZoneInfoPermanentOverlay() ZoneInfoOption {
	return zoneInfoOverlay("MANUAL", "MANUAL")
}

// ZoneInfoTimerOverlay adds a "TIMER" overlay to the tado.ZoneInfo.
//
// NOTE: timer values (DurationInSeconds, Expiry, etc) are not added to the ZoneInfo's Termination structure and will need to be added manually.
func ZoneInfoTimerOverlay() ZoneInfoOption {
	return zoneInfoOverlay("TIMER", "TIMER")
}

// ZoneInfoNextTimeBlockOverlay adds a "NEXT_TIME_BLOCK" overlay to the tado.ZoneInfo.
//
// NOTE: timer values (DurationInSeconds, Expiry, etc) are not added to the ZoneInfo's Termination structure and will need to be added manually.
func ZoneInfoNextTimeBlockOverlay() ZoneInfoOption {
	return zoneInfoOverlay("TIMER", "NEXT_TIME_BLOCK")
}
