package tado

import (
	"encoding/json"
	"fmt"
)

// Common Tado data structures

// Temperature contains a temperature in degrees Celsius
type Temperature struct {
	Celsius float64 `json:"celsius"`
}

var _ json.Marshaler = Temperature{}

// MarshalJSON implements json.Marshaler. This is needed to support SetTimeTableBlocksForDayType, since the server expects
// "null" when the temperature has not been set: {"celsius": 0} throws an error.
func (t Temperature) MarshalJSON() ([]byte, error) {
	if t.Celsius == 0 {
		return []byte(`null`), nil
	}
	return []byte(fmt.Sprintf(`{"celsius":%.1f}`, t.Celsius)), nil
}

// Percentage contains a percentage (0-100%)
type Percentage struct {
	Percentage float64 `json:"percentage"`
}

// Value contains a string value
// TODO: does this have a type as well?
type Value struct {
	Value string `json:"value"`
}

// IntValue contains an int value
type IntValue struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

// FloatValue contains a float value
type FloatValue struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}
