package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidDayTypes(t *testing.T) {
	// test that validDayTypes is valid, so we don't have to validate it at runtime
	for _, timetableID := range []TimetableID{OneDay, ThreeDay, SevenDay} {
		_, ok := validDayTypes[timetableID]
		assert.True(t, ok)
	}

}
func TestAPIClient_GetTimeTables(t *testing.T) {
	schedules := []Timetable{
		{ID: 0, Type: "ONE_DAY"},
		{ID: 1, Type: "THREE_DAY"},
		{ID: 2, Type: "SEVEN_DAY"},
	}

	c, s := makeTestServer(schedules, nil)
	defer s.Close()
	output, err := c.GetTimeTables(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, schedules, output)
}

func TestAPIClient_GetActiveTimeTable(t *testing.T) {
	active := Timetable{ID: 1, Type: "THREE_DAY"}
	c, s := makeTestServer(active, nil)
	defer s.Close()
	output, err := c.GetActiveTimeTable(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, active, output)
}

func TestAPIClient_SetActiveTimeTable(t *testing.T) {
	active := Timetable{ID: 1, Type: "THREE_DAY"}
	c, s := makeTestServer(active, nil)
	defer s.Close()
	err := c.SetActiveTimeTable(context.Background(), 1, active)
	require.NoError(t, err)
}

func TestAPIClient_GetTimeTableBlocks(t *testing.T) {
	blocks := []Block{
		{DayType: MondayToSunday, Start: "00:00", End: "07:00"},
		{DayType: MondayToSunday, Start: "07:00", End: "22:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: Temperature{Celsius: 21.0}}},
		{DayType: MondayToSunday, Start: "22:00", End: "00:00"},
	}
	c, s := makeTestServer(blocks, nil)
	defer s.Close()
	output, err := c.GetTimeTableBlocks(context.Background(), 1, 0)
	require.NoError(t, err)
	assert.Equal(t, blocks, output)

	_, err = c.GetTimeTableBlocks(context.Background(), 1, -1)
	assert.Error(t, err)
}

func TestAPIClient_GetTimeTableBlocksForDayType(t *testing.T) {
	tests := []struct {
		name        string
		timeTableID TimetableID
		dayType     DayType
		input       []Block
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:        "invalid timeTableID",
			timeTableID: 242,
			dayType:     "SATURDAY",
			wantErr:     assert.Error,
		},
		{
			name:        "invalid dayType",
			timeTableID: OneDay,
			dayType:     "SATURDAY",
			wantErr:     assert.Error,
		},
		{
			name:        "invalid dayType",
			timeTableID: ThreeDay,
			dayType:     "MONDAY",
			wantErr:     assert.Error,
		},
		{
			name:        "invalid dayType",
			timeTableID: SevenDay,
			dayType:     "foo",
			wantErr:     assert.Error,
		},
		{
			name:        "valid",
			timeTableID: OneDay,
			dayType:     "MONDAY_TO_SUNDAY",
			input: []Block{
				{DayType: "MONDAY_TO_SUNDAY", Start: "00:00", End: "07:00"},
				{DayType: "MONDAY_TO_SUNDAY", Start: "07:00", End: "22:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: Temperature{Celsius: 21.0}}},
				{DayType: "MONDAY_TO_SUNDAY", Start: "22:00", End: "00:00"},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, s := makeTestServer(tt.input, nil)
			defer s.Close()

			output, err := c.GetTimeTableBlocksForDayType(context.Background(), 1, tt.timeTableID, tt.dayType)

			assert.Equal(t, tt.input, output)
			tt.wantErr(t, err)
		})
	}
}

func TestAPIClient_SetTimeTableBlocksForDayType(t *testing.T) {
	tests := []struct {
		name        string
		timeTableID TimetableID
		dayType     DayType
		input       []Block
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:        "invalid timeTableID",
			timeTableID: 242,
			dayType:     "SATURDAY",
			wantErr:     assert.Error,
		},
		{
			name:        "invalid dayType",
			timeTableID: OneDay,
			dayType:     "SATURDAY",
			wantErr:     assert.Error,
		},
		{
			name:        "invalid dayType",
			timeTableID: ThreeDay,
			dayType:     "MONDAY",
			wantErr:     assert.Error,
		},
		{
			name:        "invalid dayType",
			timeTableID: SevenDay,
			dayType:     "foo",
			wantErr:     assert.Error,
		},
		{
			name:        "valid",
			timeTableID: OneDay,
			dayType:     "MONDAY_TO_SUNDAY",
			input: []Block{
				{DayType: "MONDAY_TO_SUNDAY", Start: "00:00", End: "07:00"},
				{DayType: "MONDAY_TO_SUNDAY", Start: "07:00", End: "22:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: Temperature{Celsius: 21.0}}},
				{DayType: "MONDAY_TO_SUNDAY", Start: "22:00", End: "00:00"},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, s := makeTestServer[[]Block](nil, nil)
			defer s.Close()

			err := c.SetTimeTableBlocksForDayType(context.Background(), 1, tt.timeTableID, tt.dayType, tt.input)
			tt.wantErr(t, err)
		})
	}
}
