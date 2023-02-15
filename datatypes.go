package tado

// Common Tado data structures

// Temperature contains a temperature in degrees Celsius
type Temperature struct {
	Celsius float64 `json:"celsius"`
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
