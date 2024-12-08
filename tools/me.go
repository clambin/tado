package tools

import (
	"context"
	"fmt"
	"github.com/clambin/tado/v2"
)

type TadoClient interface {
	GetMeWithResponse(ctx context.Context, reqEditors ...tado.RequestEditorFn) (*tado.GetMeResponse, error)
}

// GetHomes returns the list of Home ID's that are registered for the user's account.
func GetHomes(ctx context.Context, client TadoClient) ([]tado.HomeBase, error) {
	me, err := client.GetMeWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMeWithResponse: %w", err)
	}
	if me.JSON200 == nil {
		return nil, Errors(me.JSON401)
	}
	return *me.JSON200.Homes, nil
}
