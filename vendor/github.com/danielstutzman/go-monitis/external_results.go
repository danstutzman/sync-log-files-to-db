package monitis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// See http://www.monitis.com/docs/apiActions.html#getExternalMonitorResults
type GetExternalResultsOptions struct {
	// Day that results should be retrieved for.
	// Default value is the first day of the given month.
	Day *int `param:"day"`

	// Month that results should be retrieved for.
	// Default value is the current month
	Month *int `param:"month"`

	// Year that results should be retrieved for.
	// Default value is the current year
	Year *int `param:"year"`

	// Comma separated ids of locations for which results should be retrieved.
	// If not specified results will be retrieved for all locations.
	LocationIds *string `param:"locationIds"`

	// Offset relative to GMT,
	// used to show results in the timezone of the user.
	Timezone *int `param:"timezone"`

	// Default value for XML output is "HH:mm" and for JSON output is
	// "yyyy-MM-dd HH:mm". You can find some common pattern strings at
	// http://docs.oracle.com/javase/1.5.0/docs/api/java/text/SimpleDateFormat.html
	TimeFormat *string `param:"timeformat"`

	// Valid values are: last24hour, last3day, last7day, last30day.
	Period *string `param:"period"`
}

type GetExternalResultsOutput struct {
	LocationName string          `json:"locationName"` // e.g. "USA-WST"
	DataTuples   [][]interface{} `json:"data"`
	Trend        PointsTrend     `json:"trend"`
	Points       []Point
}

type Point struct {
	Timestamp time.Time
	Duration  float64
	WasOkay   bool
}

type PointsTrend struct {
	Min        float64 `json:"min"`
	OkCount    int     `json:"okcount"`
	Max        float64 `json:"max"`
	OkSum      float64 `json:"oksum"`
	NotOkCount float64 `json:"nokcount"`
}

func (auth *Auth) GetExternalResults(testId string,
	opts *GetExternalResultsOptions) ([]GetExternalResultsOutput, error) {

	client := &http.Client{}

	form := optsToForm(opts)

	form.Add("action", "testresult")
	form.Add("testId", testId)
	form.Add("apikey", auth.ApiKey)
	form.Add("authToken", auth.AuthToken)
	request, err := http.NewRequest("GET",
		"http://www.monitis.com/api?"+form.Encode(), nil)
	if err != nil {
		return []GetExternalResultsOutput{},
			fmt.Errorf("Error from NewRequest: %s", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return []GetExternalResultsOutput{},
			fmt.Errorf("Error from client.Do: %s", err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	if body[0] == '{' {
		output := map[string]interface{}{}
		err = json.Unmarshal(body, &output)
		if err != nil {
			return []GetExternalResultsOutput{},
				fmt.Errorf("Error from Unmarshal as object: %s", err)
		}
		return []GetExternalResultsOutput{},
			fmt.Errorf("API responded with error: %s", output["error"])
	} else if body[0] == '[' {
		output := []GetExternalResultsOutput{}
		err = json.Unmarshal(body, &output)
		if err != nil {
			return []GetExternalResultsOutput{},
				fmt.Errorf("Error from Unmarshal as array: %s", err)
		}

		newOutput := []GetExternalResultsOutput{}
		for _, result := range output {
			for _, tuple := range result.DataTuples {
				timestamp, err := time.Parse("2006-01-02 15:04", tuple[0].(string))
				if err != nil {
					return []GetExternalResultsOutput{},
						fmt.Errorf("Can't decode timestamp '%s': %s", tuple[0], err)
				}

				point := Point{
					Timestamp: timestamp,
					Duration:  tuple[1].(float64),
					WasOkay:   tuple[2].(string) == "OK",
				}
				result.Points = append(result.Points, point)
			}
			result.DataTuples = [][]interface{}{}
			newOutput = append(newOutput, result)
		}
		return newOutput, nil
	} else {
		return []GetExternalResultsOutput{},
			fmt.Errorf("API responded with unexpected first character: %s", body)
	}
}
