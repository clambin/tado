package tado

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// RunningTime reports the amount of time heating was on between StartTime and RunningTime, both for the home and for each individual zone.
type RunningTime struct {
	RunningTimeInSeconds int    `json:"runningTimeInSeconds"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	Zones                []struct {
		ID                   int `json:"id"`
		RunningTimeInSeconds int `json:"runningTimeInSeconds"`
	} `json:"zones"`
}

// GetRunningTimes returns the amount of time heating was on per day and per zone. from is mandatory. to is optional.
func (c *APIClient) GetRunningTimes(ctx context.Context, from, to time.Time) (runningTimes []RunningTime, err error) {
	if from.IsZero() {
		return nil, fmt.Errorf("from cannot be zero")
	}

	type response struct {
		RunningTimes []RunningTime `json:"runningTimes"`
	}
	output, err := callAPI[response](c, ctx, http.MethodGet, "minder", "/runningTimes"+buildFromToArgs(from, to), nil)
	if err == nil {
		runningTimes = output.RunningTimes
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
