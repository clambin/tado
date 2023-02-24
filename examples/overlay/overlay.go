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
