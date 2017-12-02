package monitis

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GetExternalMonitorsOutput struct {
	ExternalMonitors []ExternalMonitor `json:"testList"`
}

type ExternalMonitor struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	IsSuspended int      `json:"isSuspended"` // 1 if monitor is suspended, 0 if not
	Type        string   `json:"type"`        // e.g. "https"
	Groups      []string `json:"groups"`      // e.g. ["Default"]
	Intervals   string   `json:"intervals"`   // e.g. "15,15,15"
	Locations   string   `json:"locations"`
	Tag         string   `json:"tag"`
	Timeout     int      `json:"timeout"`
	Url         string   `json:"url"`
}

func (auth *Auth) GetExternalMonitors() ([]ExternalMonitor, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", "http://www.monitis.com/api"+
		"?action=tests"+
		"&apikey="+auth.ApiKey+
		"&authToken="+auth.AuthToken, nil)
	if err != nil {
		return []ExternalMonitor{}, fmt.Errorf("Error from NewRequest: %s", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return []ExternalMonitor{}, fmt.Errorf("Error from client.Do: %s", err)
	}

	defer response.Body.Close()
	output := GetExternalMonitorsOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return []ExternalMonitor{}, fmt.Errorf("Error from Decode: %s", err)
	}

	return output.ExternalMonitors, nil
}
