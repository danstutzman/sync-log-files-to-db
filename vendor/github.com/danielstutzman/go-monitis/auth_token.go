package monitis

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Auth struct {
	ApiKey    string
	AuthToken string
}

func GetAuthToken(apiKey, secretKey string) (*Auth, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", "http://www.monitis.com/api?action=authToken&apikey="+apiKey+"&secretkey="+secretKey, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from NewRequest: %s", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Error from client.Do: %s", err)
	}

	defer response.Body.Close()
	responseJson := map[string]string{}
	err = json.NewDecoder(response.Body).Decode(&responseJson)
	if err != nil {
		return nil, fmt.Errorf("Error from Decode: %s", err)
	}

	if responseJson["authToken"] == "" {
		return nil, fmt.Errorf("No auth token in response: %+v", responseJson)
	}

	auth := Auth{
		ApiKey:    apiKey,
		AuthToken: responseJson["authToken"],
	}
	return &auth, nil
}
