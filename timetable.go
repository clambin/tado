package tado

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/set"
	"net/http"
	"strconv"
)

// Timetable is the type of heating schedule for a Zone. Tado supports three schedule Types:
//   - ONE_DAY: same schedule for each day of the week
//   - THREE_DAY: one schedule for weekdays, one for Saturday and one for Sunday
//   - SEVEN_DAY: each day of the week has a dedicated schedule
type Timetable struct {
	ID   TimetableID `json:"id"`
	Type string      `json:"type"`
}

// TimetableID is the ID of the type of timetable
type TimetableID int

const (
	OneDay   TimetableID = 0
	ThreeDay TimetableID = 1
	SevenDay TimetableID = 2
)

// GetTimeTables returns the possible Timetable options for the provided zone
func (c *APIClient) GetTimeTables(ctx context.Context, zoneID int) (timeTables []Timetable, err error) {
	return callAPI[[]Timetable](c, ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables", nil)
}

// GetActiveTimeTable returns the active Timetable for the provided zone
func (c *APIClient) GetActiveTimeTable(ctx context.Context, zoneID int) (timeTable Timetable, err error) {
	return callAPI[Timetable](c, ctx, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/activeTimetable", nil)
}

// SetActiveTimeTable sets the active Timetable for the provided zone
func (c *APIClient) SetActiveTimeTable(ctx context.Context, zoneID int, timeTable Timetable) error {
	_, err := callAPI[struct{}](c, ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/activeTimetable", timeTable)
	return err
}

// A Block is an entry in a Timetable. It specifies the heating settings (as per the Setting attribute) for the zone at the specified time range (specified by Start and End times)
type Block struct {
	DayType             string           `json:"dayType"`
	Start               string           `json:"start"`
	End                 string           `json:"end"`
	GeolocationOverride bool             `json:"geolocationOverride"`
	Setting             ZonePowerSetting `json:"setting"`
}

// GetTimeTableBlocks returns all Block entries for a zone and timetable
func (c *APIClient) GetTimeTableBlocks(ctx context.Context, zoneID int, timetableID TimetableID) (blocks []Block, err error) {
	return callAPI[[]Block](c, ctx, http.MethodGet, "myTado",
		"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(int(timetableID))+"/blocks",
		nil)
}

// GetTimeTableBlocksForDayType returns all Block entries for a zone, timetable and day type.
//
// Valid day types are:
//   - for schedule 0 (i.e. ONE_DAY): MONDAY_TO_SUNDAY
//   - for schedule 1 (i.e. THREE_DAY): MONDAY_TO_FRIDAY, SATURDAY, SUNDAY
//   - for schedule 2 (i.e. SEVEN_DAY): MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
func (c *APIClient) GetTimeTableBlocksForDayType(ctx context.Context, zoneID int, timetableID TimetableID, dayType string) (blocks []Block, err error) {
	if err = validateDayType(timetableID, dayType); err != nil {
		return nil, err
	}
	return callAPI[[]Block](c, ctx, http.MethodGet, "myTado",
		"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(int(timetableID))+"/blocks/"+dayType,
		nil)
}

// SetTimeTableBlocksForDayType sets the Block entries for a zone, timetable and day type.
func (c *APIClient) SetTimeTableBlocksForDayType(ctx context.Context, zoneID int, timetableID TimetableID, dayType string, blocks []Block) error {
	if err := validateDayType(timetableID, dayType); err != nil {
		return err
	}
	_, err := callAPI[string](c, ctx, http.MethodPut, "myTado",
		"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(int(timetableID))+"/blocks/"+dayType,
		blocks)
	return err
}

var validDayTypes = map[TimetableID]set.Set[string]{
	OneDay:   set.Create("MONDAY_TO_SUNDAY"),
	ThreeDay: set.Create("MONDAY_TO_FRIDAY", "SATURDAY", "SUNDAY"),
	SevenDay: set.Create("MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"),
}

func validateDayType(timeTableID TimetableID, dayType string) error {
	dayTypes, ok := validDayTypes[timeTableID]
	if !ok {
		return fmt.Errorf("invalid timeTable ID: %d", timeTableID)
	}
	if !dayTypes.Contains(dayType) {
		return fmt.Errorf("invalid dayType '%s' for timeTable ID %d", dayType, timeTableID)
	}
	return nil
}
