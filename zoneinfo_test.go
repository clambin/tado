package tado_test

import (
	"context"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestAPIClient_ManualTemperature(t *testing.T) {
	server := APIServer{}
	apiServer := httptest.NewServer(http.HandlerFunc(server.apiHandler))
	defer apiServer.Close()
	authServer := httptest.NewServer(http.HandlerFunc(server.authHandler))
	defer authServer.Close()

	client := tado.APIClient{
		HTTPClient: &http.Client{},
		Username:   "user@examle.com",
		Password:   "some-password",
		AuthURL:    authServer.URL,
		APIURL:     apiServer.URL,
	}

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
