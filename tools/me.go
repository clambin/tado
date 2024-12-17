package tools

import (
	"context"
	"fmt"
	"github.com/clambin/tado/v2"
	"net/http"
)

type TadoClient interface {
	GetMeWithResponse(ctx context.Context, reqEditors ...tado.RequestEditorFn) (*tado.GetMeResponse, error)
}

// GetHomes returns the list of Home ID's that are registered under the user's account.
func GetHomes(ctx context.Context, client TadoClient) ([]tado.HomeBase, error) {
	me, err := client.GetMeWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMeWithResponse: %w", err)
	}
	if me.StatusCode() != http.StatusOK {
		return nil, HandleErrors(me.HTTPResponse, map[int]any{
			http.StatusUnauthorized: me.JSON401,
		})
	}
	return *me.JSON200.Homes, nil
}
