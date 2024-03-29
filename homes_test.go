package tado

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIClient_GetHomes(t *testing.T) {
	c, s := makeTestServer[Homes](nil, nil)
	defer s.Close()

	homes, err := c.GetHomes(context.Background())
	require.NoError(t, err)
	assert.Equal(t, Homes{{ID: 242, Name: "home"}}, homes)
}

func TestAPIClient_GetActiveHome(t *testing.T) {
	c, s := makeTestServer[Homes](nil, nil)
	defer s.Close()

	ctx := context.Background()
	home, ok := c.GetActiveHome(ctx)
	require.True(t, ok)
	assert.Equal(t, "home", home.Name)
}

func TestAPIClient_SetActiveHome(t *testing.T) {
	c, s := makeTestServer[Homes](nil, nil)
	defer s.Close()

	err := c.SetActiveHome(context.Background(), 242)
	assert.NoError(t, err)

	err = c.SetActiveHome(context.Background(), 42)
	assert.Error(t, err)
}

func TestAPIClient_SetActiveHomeByName(t *testing.T) {
	c, s := makeTestServer[Homes](nil, nil)
	defer s.Close()

	err := c.SetActiveHomeByName(context.Background(), "home")
	assert.NoError(t, err)

	err = c.SetActiveHomeByName(context.Background(), "invalid")
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
			c, s := makeTestServer(HomeState{Presence: state}, nil)
			defer s.Close()

			output, err := c.GetHomeState(context.Background())
			require.NoError(t, err)
			assert.Equal(t, state, output.Presence)
		})
	}
}

func TestAPIClient_SetHomeState(t *testing.T) {
	s := homeStateServer{HomeState: HomeState{Presence: "HOME"}}
	h := httptest.NewServer(http.HandlerFunc(s.Handle))
	defer h.Close()

	var c APIClient
	c.HTTPClient = http.DefaultClient
	c.apiURL = buildURLMap(h.URL)

	err := c.SetHomeState(context.Background(), true)
	require.NoError(t, err)
	err = c.SetHomeState(context.Background(), false)
	require.NoError(t, err)
	assert.True(t, s.HomeState.PresenceLocked)
}

func TestAPIClient_UnsetHomeState(t *testing.T) {
	s := homeStateServer{HomeState: HomeState{Presence: "HOME", PresenceLocked: true}}
	h := httptest.NewServer(http.HandlerFunc(s.Handle))
	defer h.Close()

	var c APIClient
	c.HTTPClient = http.DefaultClient
	c.apiURL = buildURLMap(h.URL)

	err := c.UnsetHomeState(context.Background())
	require.NoError(t, err)
	assert.False(t, s.HomeState.PresenceLocked)
}

type homeStateServer struct {
	HomeState
}

func (h *homeStateServer) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type not set", http.StatusUnprocessableEntity)
		return
	}

	switch r.URL.Path {
	case "/me":
		_ = json.NewEncoder(w).Encode(Account{Homes: []Home{{ID: 1}}})
	case "/homes/1/state":
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(h.HomeState)
		default:
			http.Error(w, r.URL.Path, http.StatusNotFound)
		}
	case "/homes/1/presenceLock":
		switch r.Method {
		case http.MethodPut:
			var body struct {
				HomePresence string `json:"homePresence"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			if body.HomePresence != "AWAY" && body.HomePresence != "HOME" {
				http.Error(w, body.HomePresence, http.StatusUnprocessableEntity)
				return
			}
			h.HomeState.Presence = body.HomePresence
			h.HomeState.PresenceLocked = true
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			h.HomeState.PresenceLocked = false
			w.WriteHeader(http.StatusNoContent)
		default:
		}
	default:
		log.Print(r.URL.Path)
		http.Error(w, r.URL.Path, http.StatusNotFound)
		return
	}
}

func TestHomes_FindHome(t *testing.T) {
	tests := []struct {
		name   string
		h      Homes
		lookup int
		want   Home
		want1  bool
	}{
		{
			name:   "found",
			h:      Homes{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}},
			lookup: 2,
			want:   Home{ID: 2, Name: "bar"},
			want1:  true,
		},
		{
			name:   "not found",
			h:      Homes{{ID: 1, Name: "foo"}},
			lookup: 2,
			want1:  false,
		},
		{
			name:   "empty",
			h:      Homes{},
			lookup: 2,
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.h.GetHome(tt.lookup)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func TestHomes_FindHomeByName(t *testing.T) {
	tests := []struct {
		name   string
		h      Homes
		lookup string
		want   Home
		want1  bool
	}{
		{
			name:   "found",
			h:      Homes{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}},
			lookup: "bar",
			want:   Home{ID: 2, Name: "bar"},
			want1:  true,
		},
		{
			name:   "not found",
			h:      Homes{{ID: 1, Name: "foo"}},
			lookup: "bar",
			want1:  false,
		},
		{
			name:   "empty",
			h:      Homes{},
			lookup: "bar",
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.h.GetHomeByName(tt.lookup)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
