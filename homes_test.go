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

	ctx := context.Background()
	home, ok := c.GetActiveHome(ctx)
	require.True(t, ok)
	assert.Equal(t, "home", home.Name)
}

func TestAPIClient_SetActiveHome(t *testing.T) {
	c, s := makeTestServer(nil, nil)
	defer s.Close()

	err := c.SetActiveHome(context.Background(), "home")
	assert.NoError(t, err)

	err = c.SetActiveHome(context.Background(), "invalid")
	assert.Error(t, err)
}

func TestAPIClient_GetHomeInfo(t *testing.T) {
	info := HomeInfo{Name: "my home"}
	c, s := makeTestServer(info, nil)
	defer s.Close()

	home, err := c.GetHomeInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, info, home)
}
