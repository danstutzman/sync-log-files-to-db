package monitis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// testIds: Comma separated ids of the tests to delete.
// clientApiKey: Required in case reseller is deleting client's monitor.
//   Specifies the client's API key.
func (auth *Auth) DeleteExternalMonitors(testIds string,
	clientApiKey *string) error {

	form := url.Values{}
	form.Add("action", "deleteExternalMonitor")
	form.Add("testIds", testIds)
	if clientApiKey != nil {
		form.Add("clientApiKey", *clientApiKey)
	}
	form.Add("apikey", auth.ApiKey)
	form.Add("authToken", auth.AuthToken)
	request, err := http.NewRequest("POST", "http://www.monitis.com/api",
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("Error from NewRequest: %s", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Error from client.Do: %s", err)
	}
	defer response.Body.Close()

	output := map[string]interface{}{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return fmt.Errorf("Error from Decode: %s", err)
	}

	if output["status"] != "ok" {
		return fmt.Errorf("API responded with non-OK status: %s", output)
	}

	return nil
}
