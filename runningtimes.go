package tado

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

/*
TODO: incident data structure

type Incident struct{}

	func (c *APIClient) GetIncidents(ctx context.Context) (incidents []Incident, err error) {
		if err = c.initialize(ctx); err == nil {
			var output struct {
				Incidents []Incident
			}
			if err = c.call(ctx, http.MethodGet, "minder", "/incidents", nil, &output); err == nil {
				incidents = output.Incidents
			}
		}
		return incidents, err
	}
*/

type RunningTime struct {
	RunningTimeInSeconds int    `json:"runningTimeInSeconds"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	Zones                []struct {
		Id                   int `json:"id"`
		RunningTimeInSeconds int `json:"runningTimeInSeconds"`
	} `json:"zones"`
}

// GetRunningTimes returns the amount the time heating was on per day and per zone. from is mandatory. to is optional.
func (c *APIClient) GetRunningTimes(ctx context.Context, from, to time.Time) (runningTimes []RunningTime, err error) {
	if from.IsZero() {
		return nil, fmt.Errorf("from cannot be zero")
	}
	if err = c.initialize(ctx); err == nil {
		var output struct {
			RunningTimes []RunningTime `json:"runningTimes"`
		}
		if err = c.call(ctx, http.MethodGet, "minder", "/runningTimes"+buildFromToArgs(from, to), nil, &output); err == nil {
			runningTimes = output.RunningTimes
		}
	}
	return
}

func buildFromToArgs(from, to time.Time) string {
	values := make(url.Values)
	if !from.IsZero() {
		values.Set("from", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		values.Set("to", to.Format("2006-01-02"))
	}
	encoded := values.Encode()
	if len(encoded) > 0 {
		encoded = "?" + encoded
	}
	return encoded
}
