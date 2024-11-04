package tools

import (
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   *tado.ErrorResponse
		isErr assert.ErrorAssertionFunc
		want  string
	}{
		{
			name:  "nil error",
			isErr: assert.NoError,
		},
		{
			name: "single error",
			err: &tado.ErrorResponse{Errors: &[]tado.Error{
				{Code: VarP("foo"), Title: VarP("error foo")},
			}},
			isErr: assert.Error,
			want:  "foo: error foo",
		},
		{
			name: "multiple errors",
			err: &tado.ErrorResponse{Errors: &[]tado.Error{
				{Code: VarP("foo"), Title: VarP("error foo")},
				{Code: VarP("bar"), Title: VarP("error bar")},
			}},
			isErr: assert.Error,
			want:  "foo: error foo\nbar: error bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Errors(tt.err)
			tt.isErr(t, err)
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			}
		})
	}
}

func TestErrors422(t *testing.T) {
	tests := []struct {
		name  string
		err   *tado.ErrorResponse422
		isErr assert.ErrorAssertionFunc
		want  string
	}{
		{
			name:  "nil error",
			isErr: assert.NoError,
		},
		{
			name: "single error",
			err: &tado.ErrorResponse422{Errors: &[]tado.Error422{
				{Code: VarP("foo"), Title: VarP("error foo")},
			}},
			isErr: assert.Error,
			want:  "foo: error foo",
		},
		{
			name: "multiple errors",
			err: &tado.ErrorResponse422{Errors: &[]tado.Error422{
				{Code: VarP("foo"), Title: VarP("error foo")},
				{Code: VarP("bar"), Title: VarP("error bar")},
			}},
			isErr: assert.Error,
			want:  "foo: error foo\nbar: error bar",
		},
		{
			name: "with zoneType",
			err: &tado.ErrorResponse422{Errors: &[]tado.Error422{
				{Code: VarP("foo"), Title: VarP("error foo"), ZoneType: VarP(tado.HEATING)},
			}},
			isErr: assert.Error,
			want:  "foo: error foo (zoneType: HEATING)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Errors422(tt.err)
			tt.isErr(t, err)
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			}
		})
	}
}

func VarP[T any](t T) *T {
	return &t
}
