package main

import (
	"context"
	"fmt"
	"github.com/clambin/tado"
	"os"
)

func main() {
	c, _ := tado.New(os.Getenv("TADO_USERNAME"), os.Getenv("TADO_PASSWORD"), "")
	ctx := context.Background()

	zones, err := c.GetZones(ctx)
	if err != nil {
		panic(err)
	}

	for _, zone := range zones {
		info, err := c.GetZoneInfo(ctx, zone.ID)
		if err != nil {
			panic(err)
		}

		heating := info.Setting.Power
		if heating == "ON" {
			heating = fmt.Sprintf("%.1f%%, target temperature: %.1fºC",
				info.ActivityDataPoints.HeatingPower.Percentage,
				info.Setting.Temperature.Celsius,
			)
		}
		fmt.Printf("%s: temperature: %.1fºC, humidity: %.1f%%, heating: %s\n",
			zone.Name,
			info.SensorDataPoints.InsideTemperature.Celsius,
			info.SensorDataPoints.Humidity.Percentage,
			heating,
		)
	}
}
