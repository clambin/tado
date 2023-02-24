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

	info, err := c.GetWeatherInfo(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("weather: %s\n", info.WeatherState.Value)
	fmt.Printf("temperature: %.1fºC\n", info.OutsideTemperature.Celsius)
	fmt.Printf("solar intensity: %.1f%%\n", info.SolarIntensity.Percentage)
}
