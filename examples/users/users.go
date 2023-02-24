package main

import (
	"context"
	"fmt"
	"github.com/clambin/tado"
	"os"
)

func main() {
	c := tado.New(os.Getenv("TADO_USERNAME"), os.Getenv("TADO_PASSWORD"), "")
	ctx := context.Background()

	users, err := c.GetUsers(ctx)
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		var home tado.MobileDeviceLocationState
		for _, device := range user.MobileDevices {
			home = device.IsHome()
			if home == tado.DeviceHome {
				break
			}
		}
		var status string
		switch home {
		case tado.DeviceHome:
			status = "(home)"
		case tado.DeviceAway:
			status = "(away)"
		case tado.DeviceUnknown:
		}

		fmt.Printf("%s (%s) %s\n", user.Name, user.Username, status)
	}
}
