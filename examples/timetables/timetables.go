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

	zones, err := c.GetZones(ctx)
	if err != nil {
		panic(err)
	}
	for _, zone := range zones {
		if err = showTimetable(ctx, c, zone); err != nil {
			panic(err)
		}
	}
}

func showTimetable(ctx context.Context, c *tado.APIClient, zone tado.Zone) error {
	activeTimetable, err := c.GetActiveTimeTable(ctx, zone.ID)
	if err != nil {
		return err
	}

	blocks, err := c.GetTimeTableBlocks(ctx, zone.ID, activeTimetable.ID)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		setting := block.Setting.Power
		if setting == "ON" {
			setting += fmt.Sprintf("  %.1fºC", block.Setting.Temperature.Celsius)
		}

		fmt.Printf("%-20s: %-20s: %s-%3s %s\n", zone.Name, block.DayType, block.Start, block.End, setting)
	}
	return nil
}
