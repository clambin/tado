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
			input: &APIError{Errors: []errorEntry{{Code: "foo", Title: "error1"}}},
			want:  `{"foo":"error1"}`,
		},
		{
			name:  "multiple",
			input: &APIError{Errors: []errorEntry{{Code: "foo", Title: "error1"}, {Code: "foo", Title: "error2"}}},
			want:  `[{"foo":"error1"},{"foo":"error2"}]`,
		},
		{
			name:  "empty",
			input: &APIError{},
			want:  "[]",
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

			e2 := &UnprocessableEntryError{err: tt.input}
			assert.Equal(t, "unprocessable entity: "+tt.want, e2.Error())
		})
	}
}

func TestError_Is(t *testing.T) {
	err1 := &APIError{Errors: []errorEntry{{Code: "foo", Title: "error1"}}}

	err2 := &APIError{}
	assert.ErrorIs(t, err1, err2)
	err3 := errors.New("some error")
	assert.NotErrorIs(t, err1, err3)
}

func TestErrUnprocessableEntity_Is(t *testing.T) {
	err1 := &UnprocessableEntryError{err: &APIError{Errors: []errorEntry{{Code: "foo", Title: "error1"}}}}

	err2 := &UnprocessableEntryError{}
	assert.ErrorIs(t, err1, err2)
	err3 := errors.New("some error")
	assert.NotErrorIs(t, err1, err3)
}

func TestErrUnprocessableEntry_Unwrap(t *testing.T) {
	err1 := &UnprocessableEntryError{err: &APIError{Errors: []errorEntry{{Code: "foo", Title: "error1"}}}}

	err2 := err1.Unwrap()
	var err3 *APIError
	assert.ErrorIs(t, err2, err3)

	assert.True(t, errors.As(err1, &err3))
	assert.Equal(t, `unprocessable entity: {"foo":"error1"}`, err1.Error())
}
