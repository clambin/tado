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
func GetHomes(ctx context.Context, client TadoClient) ([]tado.HomeId, error) {
	me, err := client.GetMeWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMeWithResponse: %w", err)
	}
	if me.JSON200 == nil {
		return nil, Errors(me.JSON401)
	}

	homeIds := make([]tado.HomeId, 0, len(*me.JSON200.Homes))
	for _, home := range *me.JSON200.Homes {
		homeIds = append(homeIds, *home.Id)
	}
	return homeIds, nil
}
