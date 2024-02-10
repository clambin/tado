package tado_test

import (
	"context"
	"fmt"
	"github.com/clambin/tado"
	"os"
)

func ExampleAPIClient_GetZones() {
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

func ExampleAPIClient_GetWeatherInfo() {
	c, _ := tado.New(os.Getenv("TADO_USERNAME"), os.Getenv("TADO_PASSWORD"), "")
	ctx := context.Background()

	info, err := c.GetWeatherInfo(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("weather: %s\n", info.WeatherState.Value)
	fmt.Printf("temperature: %.1fºC\n", info.OutsideTemperature.Celsius)
	fmt.Printf("solar intensity: %.1f%%\n", info.SolarIntensity.Percentage)
}

func ExampleAPIClient_GetUsers() {
	c, _ := tado.New(os.Getenv("TADO_USERNAME"), os.Getenv("TADO_PASSWORD"), "")
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

func ExampleAPIClient_GetActiveTimeTable() {
	c, _ := tado.New(os.Getenv("TADO_USERNAME"), os.Getenv("TADO_PASSWORD"), "")
	ctx := context.Background()

	zones, err := c.GetZones(ctx)
	if err != nil {
		panic(err)
	}
	for _, zone := range zones {
		activeTimetable, err := c.GetActiveTimeTable(ctx, zone.ID)
		if err != nil {
			panic(err)
		}

		blocks, err := c.GetTimeTableBlocks(ctx, zone.ID, activeTimetable.ID)
		if err != nil {
			panic(err)
		}

		for _, block := range blocks {
			setting := block.Setting.Power
			if setting == "ON" {
				setting += fmt.Sprintf("  %.1fºC", block.Setting.Temperature.Celsius)
			}

			fmt.Printf("%-20s: %-20s: %s-%3s %s\n", zone.Name, block.DayType, block.Start, block.End, setting)
		}
	}
}

func ExampleAPIClient_DeleteZoneOverlay() {
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
		if info.Overlay.GetMode() == tado.PermanentOverlay {
			if err = c.DeleteZoneOverlay(ctx, zone.ID); err != nil {
				panic(err)
			}
			fmt.Printf("removed permanent overlay from zone %s\n", zone.Name)
		}
	}
}
