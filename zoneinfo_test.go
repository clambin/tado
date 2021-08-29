package tado_test

import (
	"context"
	"github.com/clambin/tado"
	"github.com/clambin/tado/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestZoneInfo_GetState(t *testing.T) {
	zoneInfo := tado.ZoneInfo{
		Setting: tado.ZoneInfoSetting{
			Power:       "ON",
			Temperature: tado.Temperature{Celsius: 17.0},
		},
	}
	assert.Equal(t, tado.ZoneState(tado.ZoneStateAuto), zoneInfo.GetState())

	zoneInfo.Overlay = tado.ZoneInfoOverlay{
		Type: "MANUAL",
		Setting: tado.ZoneInfoOverlaySetting{
			Type:        "HEATING",
			Power:       "ON",
			Temperature: tado.Temperature{Celsius: 22.0},
		},
		Termination: tado.ZoneInfoOverlayTermination{
			Type: "MANUAL",
		},
	}
	assert.Equal(t, tado.ZoneState(tado.ZoneStateManual), zoneInfo.GetState())

	zoneInfo.Overlay.Termination.Type = "AUTO"
	assert.Equal(t, tado.ZoneState(tado.ZoneStateTemporaryManual), zoneInfo.GetState())

	zoneInfo.Overlay.Setting.Temperature.Celsius = 5.0
	assert.Equal(t, tado.ZoneState(tado.ZoneStateOff), zoneInfo.GetState())

	zoneInfo.Setting.Power = "OFF"
	zoneInfo.Setting.Temperature.Celsius = 0
	zoneInfo.Overlay = tado.ZoneInfoOverlay{}

	assert.Equal(t, tado.ZoneState(tado.ZoneStateOff), zoneInfo.GetState())
}

func TestAPIClient_ManualTemperature(t *testing.T) {
	server := APIServer{}
	apiServer := httptest.NewServer(http.HandlerFunc(server.apiHandler))
	defer apiServer.Close()
	authenticator := &mocks.Authenticator{}
	authenticator.
		On("AuthHeaders", mock.AnythingOfType("*context.emptyCtx")).
		Return(http.Header{"Authorization": []string{"Bearer good_token"}}, nil)

	client := tado.New("user@examle.com", "some-password", "")
	client.APIURL = apiServer.URL
	client.Authenticator = authenticator

	ctx := context.Background()

	err := client.SetZoneOverlay(ctx, 2, 4.0)
	assert.Nil(t, err)

	err = client.DeleteZoneOverlay(ctx, 2)
	assert.Nil(t, err)

	err = client.SetZoneOverlayWithDuration(ctx, 2, 0.0, 300*time.Second)
	assert.Nil(t, err)

	err = client.DeleteZoneOverlay(ctx, 2)
	assert.Nil(t, err)

	err = client.SetZoneOverlayWithDuration(ctx, 2, 10.0, 0)
	assert.Nil(t, err)

}
