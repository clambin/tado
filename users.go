package tado

import (
	"context"
	"net/http"
)

type User struct {
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	Username      string         `json:"username"`
	Homes         []Home         `json:"homes"`
	Locale        string         `json:"locale"`
	MobileDevices []MobileDevice `json:"mobileDevices"`
}

func (c *APIClient) GetUsers(ctx context.Context) (users []User, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "/users", nil, &users)
	}
	return
}
