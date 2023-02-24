package tado

import (
	"fmt"
	"strings"
)

var _ error = &APIError{}

// APIError contains errors received from the Tado servers when calling an API
type APIError struct {
	Errors []errorEntry `json:"errors"`
}

type errorEntry struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

func (e errorEntry) String() string {
	return fmt.Sprintf(`{"%s":"%s"}`, e.Code, e.Title)
}

// Error implements the Error interface. It returns a string representation of the error.
func (e *APIError) Error() string {
	var errors []string
	for _, entry := range e.Errors {
		errors = append(errors, entry.String())
	}
	if len(errors) == 1 {
		return errors[0]
	}
	return "[" + strings.Join(errors, ",") + "]"
}

// Is returns true if e2 is an APIError
func (e *APIError) Is(e2 error) bool {
	_, ok := e2.(*APIError)
	return ok
}

// UnprocessableEntryError indicates an API call returned http.StatusUnprocessableEntity, meaning the Tado servers
// could not parse the request
type UnprocessableEntryError struct {
	err error
}

// Error implements the Error interface. It returns a string representation of the error.
func (e *UnprocessableEntryError) Error() string {
	return "unprocessable entity: " + e.err.Error()
}

// Is returns true if e2 is an UnprocessableEntryError
func (e *UnprocessableEntryError) Is(e2 error) bool {
	_, ok := e2.(*UnprocessableEntryError)
	return ok
}

// Unwrap returns the wrapped APIError
func (e *UnprocessableEntryError) Unwrap() error {
	return e.err
}
