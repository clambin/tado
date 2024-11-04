package tools

import (
	"errors"
	"fmt"
	"github.com/clambin/tado/v2"
)

// Errors converts the errors in ErrorResponse to a Go error
func Errors(errs *tado.ErrorResponse) error {
	var err error
	if errs != nil {
		for _, e := range *errs.Errors {
			err = errors.Join(err, fmt.Errorf("%s: %s", *e.Code, *e.Title))
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
				e1 = fmt.Errorf("%s: %s", *e.Code, *e.Title)
			} else {
				err = fmt.Errorf("%s: %s (zoneType: %s)", *e.Code, *e.Title, *e.ZoneType)
			}
			err = errors.Join(err, e1)
		}
	}
	return err
}
