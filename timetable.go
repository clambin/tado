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

func (t TimetableID) isValid() bool {
	return t == OneDay || t == ThreeDay || t == SevenDay
}

const (
	// OneDay timetables have one setting for all days of the week
	OneDay TimetableID = 0
	// ThreeDay timetables have one setting for working days, one for Saturday and one for Sunday
	ThreeDay TimetableID = 1
	// SevenDay timetables have individual settings for each day of the week
	SevenDay TimetableID = 2
)

// DayType is the type of day. Valid DayType values depend on the Timetable that the block will be included in:
//   - for the OneDay timetable: MONDAY_TO_SUNDAY
//   - for the ThreeDay timetable: MONDAY_TO_FRIDAY, SATURDAY, SUNDAY
//   - for the SevenDay timetable: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
type DayType string

func (d DayType) isValid(timetableID TimetableID) bool {
	if !timetableID.isValid() {
		return false
	}
	// don't need to check if validDayTypes contains all valid timetableIDs: already done during unit testing
	dayTypes := validDayTypes[timetableID]
	return dayTypes.Contains(d)
}

var validDayTypes = map[TimetableID]set.Set[DayType]{
	OneDay:   set.Create(MondayToSunday),
	ThreeDay: set.Create(MondayToFriday, Saturday, Sunday),
	SevenDay: set.Create(Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday),
}

const (
	MondayToSunday DayType = "MONDAY_TO_SUNDAY"
	MondayToFriday DayType = "MONDAY_TO_FRIDAY"
	Monday         DayType = "MONDAY"
	Tuesday        DayType = "TUESDAY"
	Wednesday      DayType = "WEDNESDAY"
	Thursday       DayType = "THURSDAY"
	Friday         DayType = "FRIDAY"
	Saturday       DayType = "SATURDAY"
	Sunday         DayType = "SUNDAY"
)

// GetTimeTables returns the possible Timetable options for the provided zone
func (c *APIClient) GetTimeTables(ctx context.Context, zoneID int) (timeTables []Timetable, err error) {
	return callAPI[[]Timetable](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables", nil)
}

// GetActiveTimeTable returns the active Timetable for the provided zone
func (c *APIClient) GetActiveTimeTable(ctx context.Context, zoneID int) (timeTable Timetable, err error) {
	return callAPI[Timetable](ctx, c, http.MethodGet, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/activeTimetable", nil)
}

// SetActiveTimeTable sets the active Timetable for the provided zone
func (c *APIClient) SetActiveTimeTable(ctx context.Context, zoneID int, timeTable Timetable) error {
	_, err := callAPI[struct{}](ctx, c, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/schedule/activeTimetable", timeTable)
	return err
}

// A Block is an entry in a Timetable. It specifies the heating settings (as per the Setting attribute) for the zone
// at the specified DayType and time range (specified by Start and End times).
//
// The DayType must be valid for the type of timetable that it will be added to. See DayType for details.
type Block struct {
	DayType             DayType          `json:"dayType"`
	Start               string           `json:"start"`
	End                 string           `json:"end"`
	GeolocationOverride bool             `json:"geolocationOverride"`
	Setting             ZonePowerSetting `json:"setting"`
}

// GetTimeTableBlocks returns all Block entries for a zone and timetable
func (c *APIClient) GetTimeTableBlocks(ctx context.Context, zoneID int, timetableID TimetableID) (blocks []Block, err error) {
	if !timetableID.isValid() {
		return nil, fmt.Errorf("invalid timetableID")
	}
	return callAPI[[]Block](ctx, c, http.MethodGet, "myTado",
		"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(int(timetableID))+"/blocks",
		nil)
}

// GetTimeTableBlocksForDayType returns all Block entries for a zone, timetable and day type.
func (c *APIClient) GetTimeTableBlocksForDayType(ctx context.Context, zoneID int, timetableID TimetableID, dayType DayType) (blocks []Block, err error) {
	if !dayType.isValid(timetableID) {
		return nil, fmt.Errorf("invalid DayType for TimetableID %d: %s", timetableID, dayType)
	}
	return callAPI[[]Block](ctx, c, http.MethodGet, "myTado",
		"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(int(timetableID))+"/blocks/"+string(dayType),
		nil)
}

// SetTimeTableBlocksForDayType sets the Block entries for a zone, timetable and day type.
//
// The DayType must be valid for the type of timetable that it will be added to. See DayType for details.
func (c *APIClient) SetTimeTableBlocksForDayType(ctx context.Context, zoneID int, timetableID TimetableID, dayType DayType, blocks []Block) error {
	if !dayType.isValid(timetableID) {
		return fmt.Errorf("invalid DayType for TimetableID %d: %s", timetableID, dayType)
	}
	_, err := callAPI[struct{}](ctx, c, http.MethodPut, "myTado",
		"/zones/"+strconv.Itoa(zoneID)+"/schedule/timetables/"+strconv.Itoa(int(timetableID))+"/blocks/"+string(dayType),
		blocks)
	return err
}
