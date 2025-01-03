package tools

import (
	"errors"
	"github.com/clambin/tado/v2"
	"net/http"
	"testing"
)

func TestHandleErrors(t *testing.T) {
	tests := []struct {
		name     string
		resp     *http.Response
		tadoErrs map[int]any
		want     string
	}{
		{
			name: "http error",
			resp: &http.Response{StatusCode: http.StatusTooManyRequests, Status: http.StatusText(http.StatusTooManyRequests)},
			want: "http: 429 - Too Many Requests",
		},
		{
			name: "tado error",
			resp: &http.Response{StatusCode: http.StatusForbidden},
			tadoErrs: map[int]any{
				http.StatusForbidden: &tado.ErrorResponse{
					Errors: &[]tado.Error{{Code: VarP("auth"), Title: VarP("bad token")}},
				},
			},
			want: "tado: 403 - auth - bad token",
		},
		{
			name: "tado 422 error",
			resp: &http.Response{StatusCode: http.StatusUnprocessableEntity},
			tadoErrs: map[int]any{
				http.StatusUnprocessableEntity: &tado.ErrorResponse422{
					Errors: &[]tado.Error422{{
						Code:     VarP("invalid"),
						Title:    VarP("invalid value for zone type"),
						ZoneType: VarP(tado.HEATING),
					}},
				},
			},
			want: "tado: 422 - invalid - invalid value for zone type (zoneType: HEATING)",
		},
		{
			name: "Go error",
			resp: &http.Response{StatusCode: http.StatusBadRequest},
			tadoErrs: map[int]any{
				http.StatusBadRequest: errors.New("bad request"),
			},
			want: "tado: 400 - bad request",
		},
		{
			name: "not an error",
			resp: &http.Response{StatusCode: http.StatusBadGateway, Status: http.StatusText(http.StatusBadGateway)},
			tadoErrs: map[int]any{
				http.StatusBadGateway: "not an expected type",
			},
			want: "tado: 502 - Bad Gateway (string)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandleErrors(tt.resp, tt.tadoErrs).Error(); got != tt.want {
				t.Errorf("HandleErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     *tado.ErrorResponse
		wantErr bool
		want    string
	}{
		{
			name:    "nil error",
			wantErr: false,
		},
		{
			name: "single error",
			err: &tado.ErrorResponse{Errors: &[]tado.Error{
				{Code: VarP("foo"), Title: VarP("error foo")},
			}},
			wantErr: true,
			want:    "foo - error foo",
		},
		{
			name: "multiple errors",
			err: &tado.ErrorResponse{Errors: &[]tado.Error{
				{Code: VarP("foo"), Title: VarP("error foo")},
				{Code: VarP("bar"), Title: VarP("error bar")},
			}},
			wantErr: true,
			want:    "foo - error foo\nbar - error bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Errors(tt.err)
			if tt.wantErr != (err != nil) {
				t.Errorf("Errors() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if got := err.Error(); got != tt.want {
					t.Errorf("Errors() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestErrors422(t *testing.T) {
	tests := []struct {
		name    string
		err     *tado.ErrorResponse422
		wantErr bool
		want    string
	}{
		{
			name:    "nil error",
			wantErr: false,
		},
		{
			name: "single error",
			err: &tado.ErrorResponse422{Errors: &[]tado.Error422{
				{Code: VarP("foo"), Title: VarP("error foo")},
			}},
			wantErr: true,
			want:    "foo - error foo",
		},
		{
			name: "multiple errors",
			err: &tado.ErrorResponse422{Errors: &[]tado.Error422{
				{Code: VarP("foo"), Title: VarP("error foo")},
				{Code: VarP("bar"), Title: VarP("error bar")},
			}},
			wantErr: true,
			want:    "foo - error foo\nbar - error bar",
		},
		{
			name: "with zoneType",
			err: &tado.ErrorResponse422{Errors: &[]tado.Error422{
				{Code: VarP("foo"), Title: VarP("error foo"), ZoneType: VarP(tado.HEATING)},
			}},
			wantErr: true,
			want:    "foo - error foo (zoneType: HEATING)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Errors422(tt.err)
			if tt.wantErr != (err != nil) {
				t.Errorf("Errors() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if got := err.Error(); got != tt.want {
					t.Errorf("Errors422() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func VarP[T any](t T) *T {
	return &t
}
