package tado

import (
	"context"
	"fmt"
	"net/http"
)

type Account struct {
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	Username      string         `json:"username"`
	Id            string         `json:"id"`
	Homes         []Home         `json:"homes"`
	Locale        string         `json:"locale"`
	MobileDevices []MobileDevice `json:"mobileDevices"`
}

type Home struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type HomeInfo struct {
	Id                         interface{} `json:"id"`
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

func (c *APIClient) GetAccount(ctx context.Context) (Account, error) {
	var me Account
	err := c.call(ctx, http.MethodGet, "me", "", nil, &me)
	return me, err
}

func (c *APIClient) GetHomes(ctx context.Context) (homeNames []string, err error) {
	if err = c.initialize(ctx); err == nil {
		for _, home := range c.account.Homes {
			homeNames = append(homeNames, home.Name)
		}
	}
	return homeNames, err
}

func (c *APIClient) SetActiveHome(ctx context.Context, name string) (err error) {
	if err = c.initialize(ctx); err == nil {
		for _, home := range c.account.Homes {
			if home.Name == name {
				c.lock.Lock()
				c.activeHomeID = home.Id
				c.lock.Unlock()
				return nil
			}
		}
		err = fmt.Errorf("invalid home name: %s", name)
	}
	return err
}

func (c *APIClient) GetActiveHome() (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.account != nil {
		for _, home := range c.account.Homes {
			if home.Id == c.activeHomeID {
				return home.Name, true
			}
		}
	}
	return "", false
}

func (c *APIClient) GetHomeInfo(ctx context.Context) (homeInfo HomeInfo, err error) {
	if err = c.initialize(ctx); err == nil {
		err = c.call(ctx, http.MethodGet, "myTado", "", nil, &homeInfo)
	}
	return homeInfo, err
}
