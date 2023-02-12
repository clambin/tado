package tado

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIClient_GetHomes(t *testing.T) {
	c, s := makeTestServer(nil, nil)
	defer s.Close()

	homes, err := c.GetHomes(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"home"}, homes)
}

func TestAPIClient_GetActiveHome(t *testing.T) {
	c, s := makeTestServer(nil, nil)
	defer s.Close()

	_, ok := c.GetActiveHome()
	require.False(t, ok)

	_, err := c.GetHomes(context.Background())
	require.NoError(t, err)

	home, ok := c.GetActiveHome()
	require.True(t, ok)
	assert.Equal(t, "home", home)
}

func TestAPIClient_SetActiveHome(t *testing.T) {
	c, s := makeTestServer(nil, nil)
	defer s.Close()

	err := c.SetActiveHome(context.Background(), "home")
	assert.NoError(t, err)

	err = c.SetActiveHome(context.Background(), "invalid")
	assert.Error(t, err)
}
