package tado

import (
	"context"
	"fmt"
	"net/http"
)

type Account struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Id       string `json:"id"`
	Homes    []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"homes"`
	Locale        string `json:"locale"`
	MobileDevices []struct {
		Name     string `json:"name"`
		Id       int    `json:"id"`
		Settings struct {
			GeoTrackingEnabled          bool `json:"geoTrackingEnabled"`
			SpecialOffersEnabled        bool `json:"specialOffersEnabled"`
			OnDemandLogRetrievalEnabled bool `json:"onDemandLogRetrievalEnabled"`
			PushNotifications           struct {
				LowBatteryReminder          bool `json:"lowBatteryReminder"`
				AwayModeReminder            bool `json:"awayModeReminder"`
				HomeModeReminder            bool `json:"homeModeReminder"`
				OpenWindowReminder          bool `json:"openWindowReminder"`
				EnergySavingsReportReminder bool `json:"energySavingsReportReminder"`
				IncidentDetection           bool `json:"incidentDetection"`
				EnergyIqReminder            bool `json:"energyIqReminder"`
			} `json:"pushNotifications"`
		} `json:"settings"`
		DeviceMetadata struct {
			Platform  string `json:"platform"`
			OsVersion string `json:"osVersion"`
			Model     string `json:"model"`
			Locale    string `json:"locale"`
		} `json:"deviceMetadata"`
		Location struct {
			Stale           bool `json:"stale"`
			AtHome          bool `json:"atHome"`
			BearingFromHome struct {
				Degrees float64 `json:"degrees"`
				Radians float64 `json:"radians"`
			} `json:"bearingFromHome"`
			RelativeDistanceFromHomeFence float64 `json:"relativeDistanceFromHomeFence"`
		} `json:"location,omitempty"`
	} `json:"mobileDevices"`
}

func (c *APIClient) GetAccount(ctx context.Context) (Account, error) {
	var me Account
	err := c.call(ctx, http.MethodGet, "me", "", nil, &me)
	return me, err
}

func (c *APIClient) GetHomes(ctx context.Context) ([]string, error) {
	if err := c.initialize(ctx); err != nil {
		return nil, err
	}
	var homeNames []string
	for _, home := range c.account.Homes {
		homeNames = append(homeNames, home.Name)
	}
	return homeNames, nil
}

func (c *APIClient) SetActiveHome(ctx context.Context, name string) error {
	if err := c.initialize(ctx); err != nil {
		return err
	}
	for _, home := range c.account.Homes {
		if home.Name == name {
			c.lock.Lock()
			c.activeHomeID = home.Id
			c.lock.Unlock()
			return nil
		}
	}
	return fmt.Errorf("invalid home name: %s", name)
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
