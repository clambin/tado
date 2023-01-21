package tado

import (
	"context"
	"net/http"
	"time"
)

// Zone contains the response to /api/v2/homes/<HomeID>/zones
type Zone struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}

// Device contains attributes of a Tado device
type Device struct {
	DeviceType      string          `json:"deviceType"`
	Firmware        string          `json:"currentFwVersion"`
	ConnectionState ConnectionState `json:"connectionState"`
	BatteryState    string          `json:"batteryState"`
}

// ConnectionState contains the connection state of a Tado device
type ConnectionState struct {
	Value     bool      `json:"value"`
	Timestamp time.Time `json:"timeStamp"`
}

// GetZones retrieves the different Zones configured for the user's Home ID
func (client *APIClient) GetZones(ctx context.Context) (zones []Zone, err error) {
	if err = client.initialize(ctx); err != nil {
		return
	}
	err = client.call(ctx, http.MethodGet, client.apiV2URL("/zones"), nil, &zones)
	return
}
