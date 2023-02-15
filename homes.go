package tado

import (
	"context"
	"fmt"
	"net/http"
)

// Account contains details of the account used to log into the Tado API servers. Other than user id information,
// it contains all homes and all mobile devices registered under the account.
type Account struct {
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	Username      string         `json:"username"`
	ID            string         `json:"id"`
	Homes         []Home         `json:"homes"`
	Locale        string         `json:"locale"`
	MobileDevices []MobileDevice `json:"mobileDevices"`
}

// Home identifies a home registered under the Account
type Home struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetAccount returns the Account information for the account used to log into the Tado API servers.
func (c *APIClient) GetAccount(ctx context.Context) (Account, error) {
	var me Account
	err := c.call(ctx, http.MethodGet, "me", "", nil, &me)
	return me, err
}

// GetHomes returns all homes registered under the account used to log into the Tado API servers.
func (c *APIClient) GetHomes(ctx context.Context) (homeNames []string, err error) {
	if err = c.initialize(ctx); err == nil {
		for _, home := range c.account.Homes {
			homeNames = append(homeNames, home.Name)
		}
	}
	return homeNames, err
}

// SetActiveHome sets the active home for all subsequent API calls. By default, the first registered home is used.
func (c *APIClient) SetActiveHome(ctx context.Context, name string) (err error) {
	if err = c.initialize(ctx); err == nil {
		for _, home := range c.account.Homes {
			if home.Name == name {
				c.lock.Lock()
				c.activeHomeID = home.ID
				c.lock.Unlock()
				return nil
			}
		}
		err = fmt.Errorf("invalid home name: %s", name)
	}
	return err
}

// GetActiveHome returns the current active Home
func (c *APIClient) GetActiveHome(ctx context.Context) (Home, bool) {
	if err := c.initialize(ctx); err == nil {
		if c.account != nil {
			for _, home := range c.account.Homes {
				if home.ID == c.activeHomeID {
					return home, true
				}
			}
		}
	}
	return Home{}, false
}

// HomeInfo contains detailed information about a registered Home
type HomeInfo struct {
	ID                         int         `json:"id"`
	Name                       string      `json:"name"`
	DateTimeZone               string      `json:"dateTimeZone"`
	TemperatureUnit            string      `json:"temperatureUnit"`
	InstallationCompleted      bool        `json:"installationCompleted"`
	Partner                    interface{} `json:"partner"`
	SimpleSmartScheduleEnabled bool        `json:"simpleSmartScheduleEnabled"`
	ContactDetails             struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	} `json:"contactDetails"`
	Address struct {
		AddressLine1 string      `json:"addressLine1"`
		AddressLine2 interface{} `json:"addressLine2"`
		ZipCode      string      `json:"zipCode"`
		City         string      `json:"city"`
		State        interface{} `json:"state"`
		Country      string      `json:"country"`
	} `json:"address"`
	Geolocation struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"geolocation"`
}

// GetHomeInfo returns detailed information about the active Home
func (c *APIClient) GetHomeInfo(ctx context.Context) (homeInfo HomeInfo, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "", nil, &homeInfo)
	}
	return homeInfo, err
}
