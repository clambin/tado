package tools

import (
	"context"
	"fmt"
	"github.com/clambin/tado/v2"
)

func GetHomes(ctx context.Context, client *tado.ClientWithResponses) ([]tado.HomeId, error) {
	me, err := client.GetMeWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMeWithResponse: %w", err)
	}

	homeIds := make([]tado.HomeId, 0, len(*me.JSON200.Homes))
	for _, home := range *me.JSON200.Homes {
		homeIds = append(homeIds, *home.Id)
	}
	return homeIds, nil
}
