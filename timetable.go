package tado

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/clambin/go-common/set"
	"net/http"
	"strconv"
)

// TimeTable is the type of heating schedule for a Zone. Tado supports three schedule Types:
//   - ONE_DAY: same schedule for each day of the week
//   - THREE_DAY: one schedule for weekdays, one for Saturday and one for Sunday
//   - SEVEN_DAY: each day of the week has a dedicated schedule
type TimeTable struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

// GetTimeTables returns the possible TimeTable options for the provided zone
func (c *APIClient) GetTimeTables(ctx context.Context, zoneID int) (timeTables []TimeTable, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables", nil, &timeTables)
	}
	return
}

// GetActiveTimeTable returns the active Timetable for the provided zone
func (c *APIClient) GetActiveTimeTable(ctx context.Context, zoneID int) (timeTable TimeTable, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/activeTimetable", nil, &timeTable)
	}
	return
}

// SetActiveTimeTable sets the active Timetable for the provided zone
func (c *APIClient) SetActiveTimeTable(ctx context.Context, zoneID int, timeTable TimeTable) (err error) {
	if err = c.initialize(ctx); err == nil {
		buf := new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(timeTable)
		if err == nil {
			err = c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/activeTimetable", buf, nil)
		}
	}
	return
}

// A Block is an entry in a TimeTable. It specifies the heating settings (as per the Setting attribute) for the zone at the specified time range (specified by Start and End times)
type Block struct {
	DayType             string           `json:"dayType"`
	Start               string           `json:"start"`
	End                 string           `json:"end"`
	GeolocationOverride bool             `json:"geolocationOverride"`
	Setting             ZonePowerSetting `json:"setting"`
}

// GetTimeTableBlocks returns all Block entries for a zone and timetable
func (c *APIClient) GetTimeTableBlocks(ctx context.Context, zoneID int, timeTableID int) (blocks []Block, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado",
			"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(timeTableID)+"/blocks",
			nil, &blocks)
	}
	return
}

// GetTimeTableBlocksForDayType returns all Block entries for a zone, timetable and day type.
//
// Valid day types are:
//   - for schedule 0 (i.e. ONE_DAY): MONDAY_TO_SUNDAY
//   - for schedule 1 (i.e. THREE_DAY): MONDAY_TO_FRIDAY, SATURDAY, SUNDAY
//   - for schedule 2 (i.e. SEVEN_DAY): MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
func (c *APIClient) GetTimeTableBlocksForDayType(ctx context.Context, zoneID int, timeTableID int, dayType string) (blocks []Block, err error) {
	if err = validateDayType(timeTableID, dayType); err != nil {
		return nil, err
	}
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado",
			"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(timeTableID)+"/blocks/"+dayType,
			nil, &blocks)
	}
	return
}

// SetTimeTableBlocksForDayType sets the Block entries for a zone, timetable and day type.
func (c *APIClient) SetTimeTableBlocksForDayType(ctx context.Context, zoneID int, timeTableID int, dayType string, blocks []Block) (err error) {
	if err = validateDayType(timeTableID, dayType); err != nil {
		return err
	}
	if err = c.initialize(ctx); err == nil {
		buf := new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(blocks)
		if err == nil {
			err = c.call(ctx, http.MethodPut, "myTado",
				"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(timeTableID)+"/blocks/"+dayType,
				buf, nil)
		}
	}
	return
}

var validDayTypes = map[int]set.Set[string]{
	0: set.Create("MONDAY_TO_SUNDAY"),
	1: set.Create("MONDAY_TO_FRIDAY", "SATURDAY", "SUNDAY"),
	2: set.Create("MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"),
}

func validateDayType(timeTableID int, dayType string) error {
	dayTypes, ok := validDayTypes[timeTableID]
	if !ok {
		return fmt.Errorf("invalid timeTable ID: %d", timeTableID)
	}
	if !dayTypes.Contains(dayType) {
		return fmt.Errorf("invalid dayType '%s' for timeTable ID %d", dayType, timeTableID)
	}
	return nil
}
