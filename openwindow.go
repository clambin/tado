package tado

/*
import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

func (c *APIClient) SetOpenWindowDetection(ctx context.Context, zoneID int, enabled bool, timeoutInSeconds int) (openWindowDetection interface{}, err error) {
	if err = c.getActiveHomeID(ctx); err == nil {
		openWindow := struct {
			Enabled          bool `json:"enabled"`
			TimeoutInSeconds int  `json:"timeoutInSeconds"`
		}{
			Enabled:          enabled,
			TimeoutInSeconds: timeoutInSeconds,
		}
		buf := new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(openWindow)
		if err == nil {
			err = c.call(ctx, http.MethodPut, "myTado", "/zones/"+strconv.Itoa(zoneID)+"/openWindowDetection", buf, nil)
		}
	}
	return
}
*/
