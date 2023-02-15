package tado_test

import (
	"context"
	"fmt"
	"github.com/clambin/tado"
)

func ExampleZoneInfo_GetState() {
	c := tado.New("me@example.com", "password", "")
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
		fmt.Printf("%s: %s\n", zone.Name, info.GetState().String())
	}
}

func ExampleAPIClient_GetEnergySavings() {
	c := tado.New("me@example.com", "password", "")
	ctx := context.Background()

	info, err := c.GetEnergySavings(ctx)
	if err != nil {
		panic(err)
	}

	for _, report := range info {
		fmt.Printf("%s - %s: %.1f%%\n",
			report.CoveredInterval.Start,
			report.CoveredInterval.End,
			report.TotalSavings.Value,
		)
	}
}

func ExampleAPIClient_GetAirComfort() {
	c := tado.New("me@example.com", "password", "")
	ctx := context.Background()
	zones, err := c.GetZones(ctx)

	info, err := c.GetAirComfort(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("overall freshness: %s\n", info.Freshness)
	for _, room := range info.Comfort {
		if zone, ok := zones.GetZone(room.RoomID); ok {
			fmt.Printf("%s: %s / %s\n", zone.Name, room.TemperatureLevel, room.HumidityLevel)
		} else {
			fmt.Printf("unknown room: %d\n", room.RoomID)
		}
	}
}
