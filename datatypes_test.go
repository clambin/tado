package tado

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTemperature_MarshalJSON(t1 *testing.T) {
	tests := []struct {
		name        string
		temperature Temperature
		want        string
	}{
		{
			name: "empty",
			want: "null",
		},
		{
			name:        "not empty",
			temperature: Temperature{Celsius: 18.5},
			want:        `{"celsius":18.5}`,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t *testing.T) {
			got, err := tt.temperature.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
