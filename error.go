package tado

import "strings"

var _ error = &Error{}

// Error is a generic tado error
type Error struct {
	Errors []errorEntry `json:"errors"`
}

type errorEntry struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

// Error implements the error interface
func (e *Error) Error() string {
	var titles []string
	for _, entry := range e.Errors {
		titles = append(titles, entry.Title)
	}
	return strings.Join(titles, ",")
}

// Is checks if a received error is a tado Error
func (e *Error) Is(e2 error) bool {
	_, ok := e2.(*Error)
	return ok
}

// UnprocessableEntryError indicates an API call returned http.StatusUnprocessableEntity
type UnprocessableEntryError struct {
	Err error
}

// Error implements the error interface
func (e *UnprocessableEntryError) Error() string {
	return "unprocessable entity: " + e.Err.Error()
}

// Is checks if a received error is an UnprocessableEntryError
func (e *UnprocessableEntryError) Is(e2 error) bool {
	_, ok := e2.(*UnprocessableEntryError)
	return ok
}

// Unwrap returns the wrapped error
func (e *UnprocessableEntryError) Unwrap() error {
	return e.Err
}
