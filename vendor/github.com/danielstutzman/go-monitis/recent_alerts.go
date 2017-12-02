package monitis

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GetRecentAlertsOutput struct {
	Data   []RecentAlert `json:"data"`
	Status string        `json:"status"` // e.g. "ok"
}

type RecentAlert struct {
	Locs          string    `json:"locs"`         // e.g. "US-EST, US-MID2"
	DataType      string    `json:"dataType"`     // e.g. "External Monitor"
	ContactGroup  string    `json:"contactGroup"` // e.g. "All"
	AckDate       string    `json:"ackDate"`
	DataTypeId    int       `json:"dataTypeId"`
	Causes        []string  `json:"cause"`
	DataName      string    `json:"dataName"`
	RecDate       string    `json:"recDate"` // e.g. "2017-12-01 03:04:56"
	AckContact    string    `json:"ackContact"`
	DataId        int       `json:"dataId"`
	IsLocBased    bool      `json:"isLocBased"`
	Id            int       `json:"id"`
	MonitorTypeId int       `json:"monitorTypeId"`
	Contacts      []Contact `json:"contacts"`
	FailDate      string    `json:"failDate"` // e.g. "2017-12-01 02:58:26"
	MCategoryId   string    `json:"mCategoryId"`
	Status        string    `json:"status"` // e.g. "Alerted"
}

type Contact struct {
	ContactTypeId  int    `json:"contactTypeId"`
	Id             int    `json:"id"`
	ContactAccount string `json:"contactAccount"`
}

func (auth *Auth) GetRecentAlerts() ([]RecentAlert, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", "http://www.monitis.com/api"+
		"?action=recentAlerts"+
		"&apikey="+auth.ApiKey+
		"&authToken="+auth.AuthToken, nil)
	if err != nil {
		return []RecentAlert{}, fmt.Errorf("Error from NewRequest: %s", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return []RecentAlert{}, fmt.Errorf("Error from client.Do: %s", err)
	}

	defer response.Body.Close()
	output := GetRecentAlertsOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return []RecentAlert{}, fmt.Errorf("Error from Decode: %s", err)
	}

	return output.Data, nil
}
