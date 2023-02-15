package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetTimeTables(t *testing.T) {
	schedules := []Timetable{
		{ID: 0, Type: "ONE_DAY"},
		{ID: 1, Type: "THREE_DAY"},
		{ID: 2, Type: "SEVEN_DAY"},
		//{DayType: "MONDAY_TO_FRIDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "OFF"}},
		//{DayType: "SATURDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "OFF"}},
		//{DayType: "SUNDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "OFF"}},
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

/*
func TestAPIClient_GetZoneScheduleForDay(t *testing.T) {
	schedules := []Schedule{
		{DayType: "MONDAY_TO_FRIDAY", Start: "00:00", End: "00:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "OFF"}},
	}

	c, s := makeTestServer(schedules, nil)
	defer s.Close()
	output, err := c.GetZoneScheduleForDay(context.Background(), 1, "MONDAY_TO_FRIDAY")
	require.NoError(t, err)
	assert.Equal(t, schedules, output)

	err = c.SetZoneScheduleForDay(context.Background(), 1, schedules[0].DayType, schedules)
	assert.NoError(t, err)
}
*/

func TestAPIClient_GetTimeTableBlocks(t *testing.T) {
	blocks := []Block{
		{DayType: "MONDAY_TO_SUNDAY", Start: "00:00", End: "07:00"},
		{DayType: "MONDAY_TO_SUNDAY", Start: "07:00", End: "22:00", Setting: ZonePowerSetting{Type: "HEATING", Power: "ON", Temperature: Temperature{Celsius: 21.0}}},
		{DayType: "MONDAY_TO_SUNDAY", Start: "22:00", End: "00:00"},
	}
	c, s := makeTestServer(blocks, nil)
	defer s.Close()
	output, err := c.GetTimeTableBlocks(context.Background(), 1, 0)
	require.NoError(t, err)
	assert.Equal(t, blocks, output)
}

func TestAPIClient_GetTimeTableBlocksForDayType(t *testing.T) {
	tests := []struct {
		name        string
		timeTableID TimetableID
		dayType     string
		input       []Block
		pass        bool
	}{
		{
			name:        "invalid timeTableID",
			timeTableID: 242,
			dayType:     "SATURDAY",
			pass:        false,
		},
		{
			name:        "invalid dayType",
			timeTableID: OneDay,
			dayType:     "SATURDAY",
			pass:        false,
		},
		{
			name:        "invalid dayType",
			timeTableID: ThreeDay,
			dayType:     "MONDAY",
			pass:        false,
		},
		{
			name:        "invalid dayType",
			timeTableID: SevenDay,
			dayType:     "foo",
			pass:        false,
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
			pass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, s := makeTestServer(tt.input, nil)
			defer s.Close()
			output, err := c.GetTimeTableBlocksForDayType(context.Background(), 1, tt.timeTableID, tt.dayType)
			if !tt.pass {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input, output)
		})
	}
}

func TestAPIClient_SetTimeTableBlocksForDayType(t *testing.T) {
	tests := []struct {
		name        string
		timeTableID TimetableID
		dayType     string
		input       []Block
		pass        bool
	}{
		{
			name:        "invalid timeTableID",
			timeTableID: 242,
			dayType:     "SATURDAY",
			pass:        false,
		},
		{
			name:        "invalid dayType",
			timeTableID: OneDay,
			dayType:     "SATURDAY",
			pass:        false,
		},
		{
			name:        "invalid dayType",
			timeTableID: ThreeDay,
			dayType:     "MONDAY",
			pass:        false,
		},
		{
			name:        "invalid dayType",
			timeTableID: SevenDay,
			dayType:     "foo",
			pass:        false,
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
			pass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, s := makeTestServer(nil, nil)
			defer s.Close()
			err := c.SetTimeTableBlocksForDayType(context.Background(), 1, tt.timeTableID, tt.dayType, tt.input)
			if !tt.pass {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
