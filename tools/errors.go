package tools

import (
	"errors"
	"fmt"
	"github.com/clambin/tado/v2"
	"net/http"
)

// HandleErrors converts an HTTP error received from the server to a generic Go error.
// resp should be the http.Response received, and tadoErrors a map of received Error objects
// by HTTP StatusCode, for example:
//
//	me, err := client.GetMeWithResponse(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("GetMeWithResponse: %w", err)
//	}
//	if me.StatusCode() != http.StatusOK {
//		return nil, HandleErrors(me.HTTPResponse, map[int]any{
//			http.StatusUnauthorized: me.JSON401,
//		})
//	}
func HandleErrors(resp *http.Response, tadoErrors map[int]any) error {
	tadoError, ok := tadoErrors[resp.StatusCode]
	if !ok {
		return fmt.Errorf("http: %d - %s", resp.StatusCode, resp.Status)
	}
	switch e := tadoError.(type) {
	case *tado.ErrorResponse:
		return fmt.Errorf("tado: %d - %w", resp.StatusCode, Errors(e))
	case *tado.ErrorResponse422:
		return fmt.Errorf("tado: %d - %w", resp.StatusCode, Errors422(e))
	case error:
		return fmt.Errorf("tado: %d - %w", resp.StatusCode, e)
	default:
		return fmt.Errorf("tado: %d - %s (%T)", resp.StatusCode, resp.Status, tadoError)
	}
}

// Errors converts the errors in ErrorResponse to a Go error
func Errors(errs *tado.ErrorResponse) error {
	var err error
	if errs != nil {
		for _, e := range *errs.Errors {
			err = errors.Join(err, fmt.Errorf("%s - %s", *e.Code, *e.Title))
		}
	}
	return err
}

// Errors422 converts the errors in ErrorResponse422 to a Go error
func Errors422(errs *tado.ErrorResponse422) error {
	var err error
	if errs != nil {
		for _, e := range *errs.Errors {
			var e1 error
			if e.ZoneType == nil {
				e1 = fmt.Errorf("%s - %s", *e.Code, *e.Title)
			} else {
				err = fmt.Errorf("%s - %s (zoneType: %s)", *e.Code, *e.Title, *e.ZoneType)
			}
			err = errors.Join(err, e1)
		}
	}
	return err
}
