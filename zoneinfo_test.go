package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetZoneInfo(t *testing.T) {
	response := ZoneInfo{
		Setting:            ZonePowerSetting{Power: "ON", Temperature: Temperature{Celsius: 19.0}},
		ActivityDataPoints: ZoneInfoActivityDataPoints{HeatingPower: Percentage{Percentage: 75.0}},
		SensorDataPoints: ZoneInfoSensorDataPoints{
			InsideTemperature: Temperature{Celsius: 20.0},
			Humidity:          Percentage{Percentage: 10.5},
		},
		OpenWindow: ZoneInfoOpenWindow{},
		Overlay:    ZoneInfoOverlay{},
	}

	c, s := makeTestServer(response, nil)
	ctx := context.Background()
	zoneInfo, err := c.GetZoneInfo(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, response, zoneInfo)

	s.Close()
	_, err = c.GetZoneInfo(ctx, 1)
	assert.Error(t, err)
}

func TestZoneInfoOverlay_GetMode(t *testing.T) {
	type fields struct {
		Type        string
		Termination ZoneInfoOverlayTermination
	}
	tests := []struct {
		name   string
		fields fields
		want   OverlayTerminationMode
	}{
		{
			name:   "none",
			fields: fields{Type: ""},
			want:   NoOverlay,
		},
		{
			name:   "permanent",
			fields: fields{Type: "MANUAL", Termination: ZoneInfoOverlayTermination{Type: "MANUAL"}},
			want:   PermanentOverlay,
		},
		{
			name:   "timer",
			fields: fields{Type: "MANUAL", Termination: ZoneInfoOverlayTermination{Type: "TIMER", TypeSkillBasedApp: "TIMER"}},
			want:   TimerOverlay,
		},
		{
			name:   "next block",
			fields: fields{Type: "MANUAL", Termination: ZoneInfoOverlayTermination{Type: "TIMER", TypeSkillBasedApp: "NEXT_TIME_BLOCK"}},
			want:   NextBlockOverlay,
		},
		{
			name:   "unknown",
			fields: fields{Type: "TIMER", Termination: ZoneInfoOverlayTermination{Type: "bar"}},
			want:   UnknownOverlay,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := ZoneInfoOverlay{
				Type:        tt.fields.Type,
				Termination: tt.fields.Termination,
			}
			assert.Equalf(t, tt.want, z.GetMode(), "GetMode()")
		})
	}
}
