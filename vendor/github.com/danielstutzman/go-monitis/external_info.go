package monitis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GetExternalMonitorInfoOutput struct {
	// For ping monitors timeout is specified in ms.
	// For all other monitors timeout is specified in seconds.
	Timeout int `json:"timeout"`

	// Creation date of the test.
	StartDate string `json:"startDate"`

	StartDateParsed time.Time

	// One of the following test types:
	// http https ping ftp udp tcp sip smtp imap pop3 dns ssh mysql
	Type string `json:"type"`

	// This is actual if test is of type POST.
	// It is the data sent during the post request (e.g m_U=asd&m_P=asd).
	PostData string `json:"postData"`

	// integer	Id of the test
	TestId int `json:"testId"`

	// The value is 1 if there is string to match in response text
	// otherwise it is 0.
	Match int `json:"match"`

	// Text to match in the response.
	MatchText string `json:"matchText"`

	// Additional test parameters (e.g. username and password for mysql test).
	// e.g. isIPv6, isversion_1_1, sslVersion, useragent, header, sni
	Params map[string]interface{} `json:"params"`

	// first in the list group name.
	Tag string `json:"tag"`

	// Is actual for HTTP test, specifies the request method.
	// Possible values are "get", "post", "put" or "delete".
	DetailedType string `json:"detailedType"`

	// Url of the test.
	Url string `json:"url"`

	// Name of the test.
	Name string `json:"name"`

	Locations []OutputLocation `json:"locations"`

	// list of the groups the monitor belongs to
	Groups []string `json:"groups"`

	IsSuspended bool `json:"isSuspended"`

	IsPhmon bool `json:"isPhmon"`
}

type OutputLocation struct {
	// interval of checks for this location in minutes
	CheckInterval int `json:"checkInterval"`

	// full name of the location(e.g Panama, Australia, Germany ...),
	FullName string `json:"fullName"`

	// name of the location(e.g PA, AU, DE, ...),
	Name string `json:"name"`

	// id of the monitoring location.
	Id int `json:"id"`
}

func (auth *Auth) GetExternalMonitorInfo(testId string,
	timezone *int) (GetExternalMonitorInfoOutput, error) {

	client := &http.Client{}

	form := url.Values{}
	form.Add("action", "testinfo")
	form.Add("testId", testId)
	form.Add("apikey", auth.ApiKey)
	if timezone != nil {
		form.Add("timezone", strconv.Itoa(*timezone))
	}
	form.Add("authToken", auth.AuthToken)
	request, err := http.NewRequest("GET",
		"http://www.monitis.com/api?"+form.Encode(), nil)
	if err != nil {
		return GetExternalMonitorInfoOutput{},
			fmt.Errorf("Error from NewRequest: %s", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return GetExternalMonitorInfoOutput{},
			fmt.Errorf("Error from client.Do: %s", err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	output := GetExternalMonitorInfoOutput{}
	err = json.Unmarshal(body, &output)
	if err != nil {
		return GetExternalMonitorInfoOutput{},
			fmt.Errorf("Error from Unmarshal as object: %s for body %s", err, body)
	}

	output.StartDateParsed, err =
		time.Parse("01-02-2006", strings.TrimSpace(output.StartDate))
	if err != nil {
		return GetExternalMonitorInfoOutput{},
			fmt.Errorf("Can't parse StartDate '%s'", output.StartDate)
	}

	return output, nil
}
