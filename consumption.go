package tado

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Consumption contains the consumption for the total period, both in terms of how much was consumed and the associated cost
type Consumption struct {
	Currency   string `json:"currency"`
	Tariff     string `json:"tariff"`
	TariffInfo struct {
		CurrencySign    string  `json:"currencySign"`
		ConsumptionUnit string  `json:"consumptionUnit"`
		TariffInCents   float64 `json:"tariffInCents"`
		CustomTariff    bool    `json:"customTariff"`
	} `json:"tariffInfo"`
	CustomTariff          bool               `json:"customTariff"`
	ConsumptionInputState string             `json:"consumptionInputState"`
	Unit                  string             `json:"unit"`
	Details               ConsumptionDetails `json:"details"`
}

// ConsumptionDetails contains the consumption for the total period, both in terms of how much was consumed and the associated cost
type ConsumptionDetails struct {
	TotalConsumption float64             `json:"totalConsumption"`
	TotalCostInCents float64             `json:"totalCostInCents"`
	PerDay           []ConsumptionPerDay `json:"perDay"`
}

// ConsumptionPerDay contains the consumption for one day, both in terms of how much was consumed and the associated cost
type ConsumptionPerDay struct {
	Date        string  `json:"date"`
	Consumption float64 `json:"consumption"`
	CostInCents float64 `json:"costInCents"`
}

// GetConsumption returns Consumption report for the specified period. This includes both the total consumption, as well as
// the consumption per day of the period
//
// TODO:
//   - not clear what values are supported for country, or how the server uses it
//   - /consumption also supports "ngsw-bypass" (true/false) parameter, but unclear what it does
func (c *APIClient) GetConsumption(ctx context.Context, country string, start, end time.Time) (consumption Consumption, err error) {
	return callAPI[Consumption](ctx, c, http.MethodGet, "insights", "/consumption?"+buildConsumptionArgs(country, start, end).Encode(), nil)
}

func buildConsumptionArgs(country string, start, end time.Time) url.Values {
	form := make(url.Values)
	form.Set("country", country)
	if !start.IsZero() {
		form.Set("startDate", start.Format("2006-01-02"))
	}
	if !end.IsZero() {
		form.Set("endDate", end.Format("2006-01-02"))
	}
	return form
}
