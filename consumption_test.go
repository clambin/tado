package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"
)

func TestAPIClient_GetConsumption(t *testing.T) {
	info := Consumption{
		Details: ConsumptionDetails{
			TotalConsumption: 10,
			TotalCostInCents: 0,
			PerDay: []ConsumptionPerDay{
				{
					Date:        "2023-02-12",
					Consumption: 10,
					CostInCents: .4,
				},
			},
		},
	}

	c, s := makeTestServer(info, nil)
	rcvd, err := c.GetConsumption(context.Background(), "BE", time.Time{}, time.Time{})
	require.NoError(t, err)
	assert.Equal(t, info, rcvd)

	s.Close()
	_, err = c.GetConsumption(context.Background(), "BE", time.Time{}, time.Time{})
	assert.Error(t, err)

}

func Test_buildConsumptionArgs(t *testing.T) {
	type args struct {
		country string
		start   time.Time
		end     time.Time
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{
			name: "default",
			args: args{country: "BE"},
			want: url.Values{"country": []string{"BE"}},
		},
		{
			name: "start date",
			args: args{country: "BE", start: time.Date(2023, time.February, 12, 0, 0, 0, 0, time.UTC)},
			want: url.Values{"country": []string{"BE"}, "startDate": []string{"2023-02-12"}},
		},
		{
			name: "end date",
			args: args{country: "BE", end: time.Date(2023, time.February, 12, 0, 0, 0, 0, time.UTC)},
			want: url.Values{"country": []string{"BE"}, "endDate": []string{"2023-02-12"}},
		},
		{
			name: "start and end date",
			args: args{country: "BE", start: time.Date(2023, time.January, 12, 0, 0, 0, 0, time.UTC), end: time.Date(2023, time.February, 12, 0, 0, 0, 0, time.UTC)},
			want: url.Values{"country": []string{"BE"}, "startDate": []string{"2023-01-12"}, "endDate": []string{"2023-02-12"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildConsumptionArgs(tt.args.country, tt.args.start, tt.args.end))
		})
	}
}
