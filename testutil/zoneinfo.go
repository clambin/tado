package testutil

import (
	"github.com/clambin/tado"
)

func MakeZoneInfo(options ...ZoneInfoOption) tado.ZoneInfo {
	var zoneInfo tado.ZoneInfo
	for _, option := range options {
		option(&zoneInfo)
	}

	return zoneInfo
}

type ZoneInfoOption func(*tado.ZoneInfo)

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

func ZoneInfoPermanentOverlay() ZoneInfoOption {
	return zoneInfoOverlay("MANUAL", "MANUAL")
}

func ZoneInfoTimerOverlay() ZoneInfoOption {
	return zoneInfoOverlay("TIMER", "TIMER")
}

func ZoneInfoNextTimeBlockOverlay() ZoneInfoOption {
	return zoneInfoOverlay("TIMER", "NEXT_TIME_BLOCK")
}
