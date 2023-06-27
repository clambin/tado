package tado

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/clambin/tado/internal/oauth2"
	"io"
	"net/http"
	"sync"
)

// APIClient represents a Tado API client.
type APIClient struct {
	// HTTPClient is used to perform HTTP requests
	HTTPClient *http.Client

	apiURL       map[string]string
	lock         sync.RWMutex
	account      *Account
	activeHomeID int
}

// New creates a new Tado API client
//
// clientSecret can typically be left blank.  If the default secret does not work, your client secret can be found by visiting https://my.tado.com/webapp/env.js after logging in to https://my.tado.com
func New(username, password, clientSecret string) (*APIClient, error) {
	return NewWithContext(context.Background(), username, password, clientSecret)
}

// NewWithContext creates a new Tado API client for the provided Context. The API Client will no longer be valid when the Context is cancelled.
//
// clientSecret can typically be left blank.  If the default secret does not work, your client secret can be found by visiting https://my.tado.com/webapp/env.js after logging in to https://my.tado.com
func NewWithContext(ctx context.Context, username, password, clientSecret string) (*APIClient, error) {
	var client *APIClient
	httpClient, err := oauth2.NewClient(ctx, username, password, clientSecret)
	if err == nil {
		client = &APIClient{
			HTTPClient: httpClient,
			apiURL:     buildURLMap(""),
		}
	}
	return client, err
}

func buildURLMap(override string) map[string]string {
	myTado := "https://my.tado.com/api/v2"
	minder := "https://minder.tado.com/v1"
	bob := "https://energy-bob.tado.com"
	insights := "https://energy-insights.tado.com/api"

	if override != "" {
		myTado = override
		minder = override
		bob = override
		insights = override
	}

	return map[string]string{
		"me":       myTado + "/me",
		"myTado":   myTado + "/homes/%d",
		"minder":   minder + "/homes/%d",
		"bob":      bob + "/%d",
		"insights": insights + "/homes/%d",
	}
}

// callAPI is implemented as a function rather than a method, because methods cannot have type parameters (yet?)
func callAPI[T any](ctx context.Context, c *APIClient, method, apiClass, endpoint string, request any) (response T, err error) {
	if apiClass != "me" {
		if err = c.setActiveHomeID(ctx); err != nil {
			return
		}
	}

	target := c.makeAPIURL(apiClass, endpoint)
	reqBody := new(bytes.Buffer)
	if request != nil {
		if err = json.NewEncoder(reqBody).Encode(request); err != nil {
			return response, fmt.Errorf("encode: %w", err)
		}
	}

	req, _ := http.NewRequestWithContext(ctx, method, target, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, fmt.Errorf("read: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength != 0 {
			err = json.Unmarshal(respBody, &response)
		}
	case http.StatusNoContent:
	case http.StatusForbidden, http.StatusUnauthorized:
		err = errors.New(resp.Status)
	case http.StatusUnprocessableEntity:
		err = &UnprocessableEntryError{err: parseError(respBody)}
	default:
		if err = parseError(respBody); !errors.Is(err, &APIError{}) {
			err = errors.New(resp.Status)
		}
	}
	return
}

func parseError(body []byte) error {
	var errs APIError
	if err := json.Unmarshal(body, &errs); err != nil {
		return fmt.Errorf("unparsable error: %w", err)
	}
	return &errs
}

func (c *APIClient) setActiveHomeID(ctx context.Context) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.activeHomeID > 0 {
		return
	}
	account, err := c.GetAccount(ctx)
	if err != nil {
		return err
	}
	c.account = &account
	if len(c.account.Homes) == 0 {
		return fmt.Errorf("no homes detected")
	}
	c.activeHomeID = c.account.Homes[0].ID
	return nil
}

func (c *APIClient) makeAPIURL(apiClass string, endpoint string) string {
	base, ok := c.apiURL[apiClass]
	if !ok {
		panic("invalid api selector: " + base)
	}
	if apiClass == "me" {
		return base
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	return fmt.Sprintf(base, c.activeHomeID) + endpoint
}
