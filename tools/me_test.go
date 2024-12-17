package tools

import (
	"context"
	"errors"
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestGetHomes(t *testing.T) {
	tests := []struct {
		name    string
		client  fakeClient
		wantErr assert.ErrorAssertionFunc
		homes   []tado.HomeId
	}{
		{
			name: "success",
			client: fakeClient{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					JSON200:      &tado.User{Homes: &[]tado.HomeBase{{Id: VarP(tado.HomeId(1))}}},
				},
				err: nil,
			},
			wantErr: assert.NoError,
			homes:   []tado.HomeId{1},
		},
		{
			name: "failure",
			client: fakeClient{
				resp: nil,
				err:  errors.New("fail"),
			},
			wantErr: assert.Error,
		},
		{
			name: "unauthorized",
			client: fakeClient{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized},
					JSON401: &tado.Unauthorized401{
						Errors: &[]tado.Error{{Code: VarP("unauthorized"), Title: VarP("invalid credentials")}},
					},
				},
				err: nil,
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homes, err := GetHomes(context.Background(), tt.client)
			tt.wantErr(t, err)
			require.Len(t, homes, len(tt.homes))
			for i, home := range homes {
				assert.Equal(t, tt.homes[i], *home.Id)
			}
		})
	}
}

var _ TadoClient = fakeClient{}

type fakeClient struct {
	resp *tado.GetMeResponse
	err  error
}

func (f fakeClient) GetMeWithResponse(_ context.Context, _ ...tado.RequestEditorFn) (*tado.GetMeResponse, error) {
	return f.resp, f.err
}
