package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAPIClient_GetRunningTimes(t *testing.T) {
	info := struct {
		RunningTimes []RunningTime
	}{
		RunningTimes: []RunningTime{
			{RunningTimeInSeconds: 10, StartTime: "2023-02-13", EndTime: "2023-02-14"},
			{RunningTimeInSeconds: 20, StartTime: "2023-02-14", EndTime: "2023-02-15"},
		},
	}

	c, s := makeTestServer(info, nil)
	rcvd, err := c.GetRunningTimes(context.Background(), time.Now(), time.Time{})
	require.NoError(t, err)
	assert.Equal(t, info.RunningTimes, rcvd)

	_, err = c.GetRunningTimes(context.Background(), time.Time{}, time.Time{})
	assert.Error(t, err)

	s.Close()
	_, err = c.GetRunningTimes(context.Background(), time.Now(), time.Time{})
	assert.Error(t, err)
}

func Test_buildFromToArgs(t *testing.T) {
	type args struct {
		from time.Time
		to   time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no args",
			args: args{},
			want: "",
		},
		{
			name: "from",
			args: args{from: time.Date(2023, time.February, 14, 0, 0, 0, 0, time.UTC)},
			want: "?from=2023-02-14",
		},
		{
			name: "to",
			args: args{to: time.Date(2023, time.February, 14, 0, 0, 0, 0, time.UTC)},
			want: "?to=2023-02-14",
		},
		{
			name: "both",
			args: args{from: time.Date(2023, time.February, 13, 0, 0, 0, 0, time.UTC), to: time.Date(2023, time.February, 14, 0, 0, 0, 0, time.UTC)},
			want: "?from=2023-02-13&to=2023-02-14",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildFromToArgs(tt.args.from, tt.args.to))
		})
	}
}
