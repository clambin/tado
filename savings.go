package tado

import (
	"context"
	"net/http"
	"time"
)

type EnergySavingsReport struct {
	CoveredInterval struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"coveredInterval"`
	TotalSavingsAvailable bool `json:"totalSavingsAvailable"`
	WithAutoAssist        struct {
		DetectedAwayDuration     IntValue `json:"detectedAwayDuration"`
		OpenWindowDetectionTimes int      `json:"openWindowDetectionTimes"`
	} `json:"withAutoAssist"`
	TotalSavingsInThermostaticMode IntValue   `json:"totalSavingsInThermostaticMode"`
	ManualControlSaving            FloatValue `json:"manualControlSaving"`
	TotalSavings                   FloatValue `json:"totalSavings"`
	HideSunshineDuration           bool       `json:"hideSunshineDuration"`
	AwayDuration                   IntValue   `json:"awayDuration"`
	ShowSavingsInThermostaticMode  bool       `json:"showSavingsInThermostaticMode"`
	CommunityNews                  *struct {
		Type                   string `json:"type,omitempty"`
		HumidityLevelDurations struct {
			TooLow  FloatValue `json:"tooLow"`
			Optimum FloatValue `json:"optimum"`
			TooHigh FloatValue `json:"tooHigh"`
		} `json:"humidityLevelDurations,omitempty"`
		OpenWindowComparison struct {
			UserDifference int `json:"userDifference"`
			AreaAverage    int `json:"areaAverage"`
		} `json:"openWindowComparison,omitempty"`
		States []struct {
			Name  string  `json:"name"`
			Value float64 `json:"value"`
			Unit  string  `json:"unit"`
		} `json:"states,omitempty"`
		AverageTotalSavings     FloatValue `json:"averageTotalSavings,omitempty"`
		HomeCountry             string     `json:"homeCountry,omitempty"`
		AverageNightTemperature struct {
			IndoorInCelsius  float64 `json:"indoorInCelsius"`
			OutdoorInCelsius float64 `json:"outdoorInCelsius"`
		} `json:"averageNightTemperature,omitempty"`
		Value                                     string  `json:"value,omitempty"`
		AreaAverageTemperatureInCelsius           float64 `json:"areaAverageTemperatureInCelsius,omitempty"`
		HomeAverageTemperatureInCelsius           float64 `json:"homeAverageTemperatureInCelsius,omitempty"`
		TurnOnDateForMajorityOfTadoUsers          string  `json:"turnOnDateForMajorityOfTadoUsers,omitempty"`
		TurnOnDateForMajorityOfUsersInLocalRegion string  `json:"turnOnDateForMajorityOfUsersInLocalRegion,omitempty"`
	} `json:"communityNews"`
	SunshineDuration                        IntValue   `json:"sunshineDuration"`
	HasAutoAssist                           bool       `json:"hasAutoAssist"`
	OpenWindowDetectionTimes                int        `json:"openWindowDetectionTimes"`
	SetbackScheduleDurationPerDay           FloatValue `json:"setbackScheduleDurationPerDay"`
	TotalSavingsInThermostaticModeAvailable bool       `json:"totalSavingsInThermostaticModeAvailable"`
	YearMonth                               string     `json:"yearMonth"`
	HideOpenWindowDetection                 bool       `json:"hideOpenWindowDetection"`
	Home                                    int        `json:"home"`
}

func (c *APIClient) GetEnergySavings(ctx context.Context) (reports []EnergySavingsReport, err error) {
	var output struct {
		Reports []EnergySavingsReport `json:"reports"`
	}
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "bob", "/", nil, &output)
	}
	return output.Reports, err
}
