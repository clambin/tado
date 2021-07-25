package tado_test

import (
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZoneInfo_GetState(t *testing.T) {
	zoneInfo := tado.ZoneInfo{}
	assert.Equal(t, tado.ZoneStateAuto, int(zoneInfo.GetState()))

	zoneInfo.Overlay = tado.ZoneInfoOverlay{
		Type: "MANUAL",
		Setting: tado.ZoneInfoOverlaySetting{
			Type:        "HEATING",
			Temperature: tado.Temperature{Celsius: 22.0},
		},
		Termination: tado.ZoneInfoOverlayTermination{
			Type: "MANUAL",
		},
	}
	assert.Equal(t, tado.ZoneStateManual, int(zoneInfo.GetState()))

	zoneInfo.Overlay.Setting.Temperature.Celsius = 5.0
	assert.Equal(t, tado.ZoneStateOff, int(zoneInfo.GetState()))

	zoneInfo.Overlay.Termination.Type = "AUTO"
	assert.Equal(t, tado.ZoneStateTemporaryManual, int(zoneInfo.GetState()))

}
