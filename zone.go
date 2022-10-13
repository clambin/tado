package tado

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

// String serializes a Zone into a string. Used for logging
func (zone Zone) String() string {
	devicesAsStr := make([]string, len(zone.Devices))
	for i, device := range zone.Devices {
		devicesAsStr[i] = device.String()
	}
	devicesStr := strings.Join(devicesAsStr, ", ")

	return fmt.Sprintf("id=%d name=%s devices={%s}",
		zone.ID,
		zone.Name,
		devicesStr,
	)
}

// String serializes a Device into a string. Used for logging
func (device *Device) String() string {
	return fmt.Sprintf("type=%s firmware=%s connection=%v battery=%s",
		device.DeviceType,
		device.Firmware,
		device.ConnectionState.Value,
		device.BatteryState,
	)
}
