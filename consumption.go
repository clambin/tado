package tado

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

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

type ConsumptionDetails struct {
	TotalConsumption float64             `json:"totalConsumption"`
	TotalCostInCents float64             `json:"totalCostInCents"`
	PerDay           []ConsumptionPerDay `json:"perDay"`
}

type ConsumptionPerDay struct {
	Date        string  `json:"date"`
	Consumption float64 `json:"consumption"`
	CostInCents float64 `json:"costInCents"`
}

func (c *APIClient) GetConsumption(ctx context.Context, country string, start, end time.Time) (consumption Consumption, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "insights", "/consumption?"+buildConsumptionArgs(country, start, end).Encode(), nil, &consumption)
	}
	return
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
