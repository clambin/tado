package tado

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIClient_GetUsers(t *testing.T) {
	info := []User{
		{Name: "foo", Username: "foo", Homes: []Home{{ID: 1, Name: "snafu"}}, Locale: "", MobileDevices: nil},
		{Name: "bar", Username: "bar", Homes: []Home{{ID: 1, Name: "snafu"}}, Locale: "", MobileDevices: nil},
	}
	c, s := makeTestServer(info, nil)
	rcvd, err := c.GetUsers(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info, rcvd)

	s.Close()
	_, err = c.GetUsers(context.Background())
	assert.Error(t, err)
}
