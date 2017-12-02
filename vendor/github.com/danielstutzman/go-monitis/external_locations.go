package monitis

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ExternalLocation struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`        // e.g. US-MID
	DnsName     string `json:"dnsName"`     // e.g. dallas1up.monitis.com
	FullName    string `json:"fullName"`    // e.g. US-MID
	HostAddress string `json:"hostAddress"` // e.g. 75.126.39.114

	// Minimum check interval available for external monitors in minutes.
	// Possible values are 1,3,5,10,15,20,30,40,60.
	MinCheckInterval int `json:"minCheckInterval"`

	Group string `json:"group"` // e.g. Americas
	City  string `json:"city"`
}

func (auth *Auth) GetLocations() ([]ExternalLocation, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", "http://www.monitis.com/api"+
		"?action=locations"+
		"&apikey="+auth.ApiKey+
		"&authToken="+auth.AuthToken, nil)
	if err != nil {
		return []ExternalLocation{}, fmt.Errorf("Error from NewRequest: %s", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return []ExternalLocation{}, fmt.Errorf("Error from client.Do: %s", err)
	}

	defer response.Body.Close()
	output := []ExternalLocation{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return []ExternalLocation{}, fmt.Errorf("Error from Decode: %s", err)
	}

	return output, nil
}
