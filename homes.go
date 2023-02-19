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
	return callAPI[Account](ctx, c, http.MethodGet, "me", "", nil)
}

// GetHomes returns all homes registered under the account used to log into the Tado API servers.
func (c *APIClient) GetHomes(ctx context.Context) (homeNames []string, err error) {
	if err = c.getActiveHomeID(ctx); err == nil {
		for _, home := range c.account.Homes {
			homeNames = append(homeNames, home.Name)
		}
	}
	return homeNames, err
}

// SetActiveHome sets the active home for all subsequent API calls. By default, the first registered home is used.
func (c *APIClient) SetActiveHome(ctx context.Context, name string) (err error) {
	if err = c.getActiveHomeID(ctx); err == nil {
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
	if err := c.getActiveHomeID(ctx); err == nil {
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
	return callAPI[HomeInfo](ctx, c, http.MethodGet, "myTado", "", nil)
}

// HomeState contains the home state (HOME/AWAY)
type HomeState struct {
	Presence       string `json:"presence"`
	PresenceLocked bool   `json:"presenceLocked"`
}

// GetHomeState returns the home state (HOME/AWAY)
func (c *APIClient) GetHomeState(ctx context.Context) (homeState HomeState, err error) {
	return callAPI[HomeState](ctx, c, http.MethodGet, "myTado", "/state", nil)
}

// SetHomeState sets the home state (HOME/AWAY)
func (c *APIClient) SetHomeState(ctx context.Context, home bool) error {
	var homeState struct {
		HomePresence string `json:"homePresence"`
	}
	if home {
		homeState.HomePresence = "HOME"
	} else {
		homeState.HomePresence = "AWAY"
	}
	_, err := callAPI[string](ctx, c, http.MethodPut, "myTado", "/presenceLock", homeState)
	return err
}

// UnsetHomeState removes any manual presence set by SetHomeState
func (c *APIClient) UnsetHomeState(ctx context.Context) error {
	_, err := callAPI[string](ctx, c, http.MethodDelete, "myTado", "/presenceLock", nil)
	return err
}
