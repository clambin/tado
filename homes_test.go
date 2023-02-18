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

func TestAPIClient_GetHomeState(t *testing.T) {
	for _, state := range []string{"HOME", "AWAY"} {
		t.Run(state, func(t *testing.T) {
			info := HomeState{
				Presence: state,
			}
			c, s := makeTestServer(info, nil)
			defer s.Close()

			output, err := c.GetHomeState(context.Background())
			require.NoError(t, err)
			assert.Equal(t, info, output)
		})
	}
}

func TestAPIClient_SetHomeState(t *testing.T) {
	c, s := makeTestServer(nil, nil)
	defer s.Close()

	err := c.SetHomeState(context.Background(), true)
	require.NoError(t, err)
	err = c.SetHomeState(context.Background(), false)
	require.NoError(t, err)
}

func TestAPIClient_UnsetHomeState(t *testing.T) {
	c, s := makeTestServer(nil, nil)
	defer s.Close()

	err := c.UnsetHomeState(context.Background())
	require.NoError(t, err)
}
