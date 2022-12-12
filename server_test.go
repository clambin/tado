package tado_test

import (
	"context"
	"net/http"
	"time"
)

// APIServer implements an authenticating API server
type APIServer struct {
	fail bool
	slow bool
}

func (apiServer *APIServer) apiHandler(w http.ResponseWriter, req *http.Request) {
	if apiServer.fail {
		http.Error(w, "server is having issues", http.StatusInternalServerError)
		return
	}

	if apiServer.slow && wait(req.Context(), 5*time.Second) == false {
		http.Error(w, "context exceeded", http.StatusRequestTimeout)
		return
	}

	token := req.Header.Get("Authorization")
	if token != "Bearer good_token" {
		http.Error(w, "request denied", http.StatusForbidden)
		return
	}

	contentType, ok := req.Header["Content-Type"]
	if ok == false || contentType[0] != "application/json;charset=UTF-8" {
		http.Error(w, "content type should be application/json", http.StatusUnprocessableEntity)
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
