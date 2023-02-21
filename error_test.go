package tado

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name  string
		input error
		want  string
	}{
		{
			name:  "single",
			input: &Error{Errors: []errorEntry{{Code: "foo", Title: "error1"}}},
			want:  "error1",
		},
		{
			name:  "multiple",
			input: &Error{Errors: []errorEntry{{Code: "foo", Title: "error1"}, {Code: "foo", Title: "error2"}}},
			want:  "error1,error2",
		},
		{
			name:  "empty",
			input: &Error{},
			want:  "",
		},
		{
			name:  "other error",
			input: errors.New("some error"),
			want:  "some error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.input.Error())

			e2 := &ErrUnprocessableEntry{Err: tt.input}
			assert.Equal(t, "unprocessable entity: "+tt.want, e2.Error())
		})
	}
}

func TestError_Is(t *testing.T) {
	err1 := &Error{Errors: []errorEntry{{Code: "foo", Title: "error1"}}}

	err2 := &Error{}
	assert.ErrorIs(t, err1, err2)
	err3 := errors.New("some error")
	assert.NotErrorIs(t, err1, err3)
}

func TestErrUnprocessableEntity_Is(t *testing.T) {
	err1 := &ErrUnprocessableEntry{Err: &Error{Errors: []errorEntry{{Code: "foo", Title: "error1"}}}}

	err2 := &ErrUnprocessableEntry{}
	assert.ErrorIs(t, err1, err2)
	err3 := errors.New("some error")
	assert.NotErrorIs(t, err1, err3)
}

func TestErrUnprocessableEntry_Unwrap(t *testing.T) {
	err1 := &ErrUnprocessableEntry{Err: &Error{Errors: []errorEntry{{Code: "foo", Title: "error1"}}}}

	err2 := err1.Unwrap()
	var err3 *Error
	assert.ErrorIs(t, err2, err3)

	assert.True(t, errors.As(err1, &err3))
	assert.Equal(t, "unprocessable entity: error1", err1.Error())
}
