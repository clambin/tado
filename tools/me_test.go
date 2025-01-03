package tools

import (
	"context"
	"errors"
	"github.com/clambin/tado/v2"
	"net/http"
	"testing"
)

func TestGetHomes(t *testing.T) {
	tests := []struct {
		name    string
		client  fakeClient
		wantErr bool
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
			wantErr: false,
			homes:   []tado.HomeId{1},
		},
		{
			name: "failure",
			client: fakeClient{
				resp: nil,
				err:  errors.New("fail"),
			},
			wantErr: true,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homes, err := GetHomes(context.Background(), tt.client)
			if tt.wantErr != (err != nil) {
				t.Errorf("Errors() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(tt.homes) != len(homes) {
				t.Fatalf("got %d homes, expected %d", len(homes), len(tt.homes))
			}
			for i, home := range homes {
				if *home.Id != tt.homes[i] {
					t.Errorf("got home %+v, expected home %+v", *home.Id, tt.homes[i])
				}
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
