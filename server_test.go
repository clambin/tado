package tado_test

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// APIServer implements an authenticating API server
type APIServer struct {
	counter      int
	accessToken  string
	refreshToken string
	expires      time.Time
	failRefresh  bool
	slow         bool
}

func (apiServer *APIServer) authHandler(w http.ResponseWriter, req *http.Request) {
	log.Debug("authHandler: " + req.URL.Path)

	if req.URL.Path != "/oauth/token" {
		http.Error(w, "endpoint not implemented", http.StatusNotFound)
		return
	}

	response, ok := apiServer.handleAuthentication(req)

	if ok == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
	} else {
		_, _ = w.Write([]byte(response))
	}
}

func (apiServer *APIServer) apiHandler(w http.ResponseWriter, req *http.Request) {
	log.Debug("apiHandler: " + html.EscapeString(req.URL.Path))

	if apiServer.slow && wait(req.Context(), 5*time.Second) == false {
		http.Error(w, "context exceeded", http.StatusRequestTimeout)
		return
	}

	contentType, ok := req.Header["Content-Type"]
	if ok == false || contentType[0] != "application/json;charset=UTF-8" {
		http.Error(w, "content type should be application/json", http.StatusUnprocessableEntity)
		return
	}

	if apiServer.authenticateRequest(req) == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	response, ok := responses[req.URL.Path]

	if ok {
		if response == "" {
			http.Error(w, "", http.StatusNoContent)
		} else {
			_, _ = w.Write([]byte(response))
		}
	} else {
		http.Error(w, "endpoint not implemented: "+req.URL.Path, http.StatusForbidden)
	}
}

func wait(ctx context.Context, duration time.Duration) (passed bool) {
	timer := time.NewTimer(duration)
loop:
	for {
		select {
		case <-timer.C:
			break loop
		case <-ctx.Done():
			return false
		}
	}
	return true
}

func (apiServer *APIServer) handleAuthentication(req *http.Request) (response string, ok bool) {
	const authResponse = `{
  		"access_token":"%s",
  		"token_type":"bearer",
  		"refresh_token":"%s",
  		"expires_in":%d,
  		"scope":"home.user",
  		"jti":"jti"
	}`

	grantType := getGrantType(req.Body)

	if grantType == "refresh_token" {
		if apiServer.failRefresh {
			return "test server in failRefresh mode", false
		}
		apiServer.counter++
	} else {
		apiServer.counter = 1
	}

	apiServer.accessToken = fmt.Sprintf("token_%d", apiServer.counter)
	apiServer.refreshToken = apiServer.accessToken
	apiServer.expires = time.Now().Add(20 * time.Second)

	return fmt.Sprintf(authResponse, apiServer.accessToken, apiServer.refreshToken, 20), true
}

func getGrantType(body io.Reader) string {
	content, _ := ioutil.ReadAll(body)
	if params, err := url.ParseQuery(string(content)); err == nil {
		if tokenType, ok := params["grant_type"]; ok == true {
			return tokenType[0]
		}
	}
	panic("grant_type not found in body")
}

func (apiServer *APIServer) authenticateRequest(req *http.Request) (ok bool) {
	if apiServer.accessToken == "" {
		return false
	}

	bearer := req.Header.Get("Authorization")
	if bearer != "Bearer "+apiServer.accessToken {
		return false
	}

	if time.Now().After(apiServer.expires) {
		return false
	}

	return true
}

var responses = map[string]string{
	"/oauth/token": `{
  "access_token":"access_token",
  "token_type":"bearer",
  "refresh_token":"refresh_token",
  "expires_in":599,
  "scope":"home.user",
  "jti":"jti"
}`,
	"/api/v1/me": `{
  "name":"Some User",
  "email":"user@example.com",
  "username":"user@example.com",
  "enabled":true,
  "id":"somelongidstring",
  "homeId":242,
  "locale":"en_BE",
  "type":"WEB_USER"
}`,
	"/api/v2/homes/242/zones": `[
  { 
    "id": 1, 
    "name": "Living room", 
    "devices": [ 
		{
		  "deviceType": "RU02",
		  "currentFwVersion": "67.2", 
		  "connectionState": { 
			"value": true 
		  }, 
		  "batteryState": "NORMAL" 
		}
    ]
  },
  { "id": 2, "name": "Study" },
  { "id": 3, "name": "Bathroom" }
]`,
	"/api/v2/homes/242/zones/1/state": `{
  "setting": {
    "power": "ON",
    "temperature": { "celsius": 20.00 }
  },
  "openWindow": null,
  "activityDataPoints": { "heatingPower": { "percentage": 11.00 } },
  "sensorDataPoints": {
    "insideTemperature": { "celsius": 19.94 },
    "humidity": { "percentage": 37.70 }
  }
}`,
	"/api/v2/homes/242/zones/2/state": `{
  "setting": {
    "power": "ON",
    "temperature": { "celsius": 20.00 }
  },
  "openWindow": {
    "durationInSeconds": 50,
    "remainingTimeInSeconds": 250
  },
  "activityDataPoints": { "heatingPower": { "percentage": 11.00 } },
  "sensorDataPoints": {
    "insideTemperature": { "celsius": 19.94 },
    "humidity": { "percentage": 37.70 }
  }
}`,
	//type ZoneInfoOpenWindow struct {
	//	DetectedTime           time.Time `json:"detectedTime"`
	//	DurationInSeconds      int       `json:"durationInSeconds"`
	//	Delay                 time.Time `json:"expiry"`
	//	RemainingTimeInSeconds int       `json:"remainingTimeInSeconds"`
	//}
	"/api/v2/homes/242/weather": `{
  "outsideTemperature": { "celsius": 3.4 },
  "solarIntensity": { "percentage": 13.3 },
  "weatherState": { "value": "CLOUDY_MOSTLY" }
}`,
	"/api/v2/homes/242/mobileDevices": `[{
	"id": 1,
	"name": "device 1",
	"settings": {
		"geoTrackingEnabled": true
	},
	"location": {
		"stale": false,
		"atHome": true
	}
}, {
	"id": 2,
	"name": "device 2",
	"settings": {
		"geoTrackingEnabled": true
	},
	"location": {
		"stale": false,
		"atHome": false
	}
}]`,
	// TODO: this doesn't test whether PUT/DELETE were used, nor validates the payload
	"/api/v2/homes/242/zones/2/overlay": ``,
}
