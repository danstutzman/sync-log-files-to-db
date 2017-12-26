package monitis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// See http://www.monitis.com/docs/apiActions.html#addExternalMonitor
type AddExternalMonitorOptions struct {
	// One of: http https ftp ping ssh dns mysql udp tcp sip smtp imap
	//         pop certificate
	Type *string `param:"type"`

	// 1=GET 2=POST 3=PUT 4=DELETE, default 1
	DetailedTestType *int `param:"detailedTestType"`

	// Name of the test
	Name *string `param:"name"`

	// Url to test.  Must not contain protocol
	Url *string `param:"url"`

	// Test every 1, 3, or 15 minutes
	Interval *int `param:"interval"`

	// For ping monitors timeout should be specified in ms, max value is 5000. For all other monitors timeout should be specified in seconds, max value is 50.
	Timeout *int `param:"timeout"`

	// Comma separated ids of the locations to add test for.
	// Specify at least two.
	LocationIds *string `param:"locationIds"`

	// the group name or list of the goups the monitor will belong to:
	// E.g. ["dev", "ops"]
	// Can be "Default"
	Tag *string `param:"tag"`

	// If 1, requests will be sent via SSL.
	// Can be set for FTP, UDP, TCP, SMTP, IMAP, POP test types.
	// Default value is 1 for https and ssh, otherwise 0.
	OverSsl *int `param:"overSSL"`

	// Data to send during POST request, e.g m_U=asd&m_P=asd.
	PostData *string `param:"postData"`

	// Text to match in the response. If the contentMatchFlag is 1 or 2 then
	// it's one not encoded string. For contentMatchFlag 3-6 the strings should
	// be encoded with Unicode and separated with commas.
	ContentMatchString *string `param:"contentMatchString"`

	// Options for the string search:
	// Flag value:
	// 1 - The response should contain the contentMatchString string
	// 0 - The response should NOT contain the contentMatchString string
	// 3 - The response should contain ALL OF the contentMatchString strings,
	// and the strings must be encoded with Unicode and separated with comma
	// 4 - The response should contain AT LEAST ONE OF the contentMatchString
	// strings, and the strings must be encoded with Unicode and separated with
	// comma
	// 5 - The response should NOT contain ALL OF the contentMatchString strings,
	// and the strings must be encoded with Unicode and separated with comma
	// 6 - The response should NOT contain AT LEAST ONE OF the
	// contentMatchString strings, and the strings must be encoded with Unicode
	// and separated with commas.
	ContentMatchFlag *int `param:"contentMatchFlag"`

	// This parameter makes sence only in HTTP and HTTPS monitors.
	// Required additional parameters for DNS, Certificate and MySQL monitors.
	// Use the following format - key1:value1;key2:value2; . . .
	// For MySQL test:
	//   username - user name for authentication,
	//   password - password for authentication,
	//   port - MySQL port number,
	//   timeout - MySQL timeout in seconds.
	// For DNS test:
	//   server - the name server,
	//   expip - expected IP,
	//   expauth - if name server is authoritative should be "-A" otherwise an empty string.
	// For Certificate test:
	//   days - expiration days
	Params *string `param:"params"`

	// Minimal allowed uptime(%).
	UptimeSla *int `param:"uptimeSLA"`

	// Maximal allowed response time in seconds.
	ResponseSla *int `param:"responseSLA"`

	// Username for authentication.
	BasicAuthUser *string `param:"basicAuthUser"`

	// Password for authentication.
	BasicAuthPass *string `param:"basicAuthPass"`

	// List of headers used to make the http request. Format is the following:
	//   headerName1:headerValue1;headerName2:headerValue1;...
	// where headerName1, headerName2 are header names and headerValue1,
	// headerValue1 are header values.
	// This parameter makes sence only in HTTP and HTTPS monitors.
	Header *string `param:"header"`

	// Used for HTTPS and SOAP.
	// Use server Name Indication. Available values are 0, 1.
	Sni *int `param:"sni"`

	// Used for HTTP, HTTPS and SOAP.
	// Force to use HTTP version 1.1. Available values are 0, 1.
	IsVersion_1_1 *int `param:"isversion_1_1"`

	// Used for HTTP, HTTPS and SOAP.
	UserAgent *string `param:"userAgent"`

	// Specifies id of the order in scope of which monitor will be added.
	// This is required parameter for reseller's client accounts.
	OrderId *int `param:"orderId"`

	// 1 if enabled using IPv6, else 0
	IsIpv6 *int `param:"isIPv6"`
}

type AddExternalMonitorOutput struct {
	Error string `json:"error"`

	// Request status
	// (in case of something goes wrong, will contain error massage).
	// or "ok" if okay
	Status string `json:"status"`

	Data AddExternalMonitorOutputData `json:"data"`
}

type AddExternalMonitorOutputData struct {
	// Id of the test.
	TestId int `json:testId`

	//Creation date of the test.
	// e.g. "2017-12-01 12:14:0"
	StartDate string `json:startDate`

	// "0" or "1". If "1", test is new created.
	IsTestNew string `json:isTestNew`
}

func (auth *Auth) AddExternalMonitor(
	opts *AddExternalMonitorOptions) (*AddExternalMonitorOutputData, error) {

	form := optsToForm(opts)

	form.Add("action", "addExternalMonitor")
	form.Add("apikey", auth.ApiKey)
	form.Add("authToken", auth.AuthToken)
	request, err := http.NewRequest("POST", "http://www.monitis.com/api",
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("Error from NewRequest: %s", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Error from client.Do: %s", err)
	}
	defer response.Body.Close()

	output := AddExternalMonitorOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return nil, fmt.Errorf("Error from Decode: %s", err)
	}

	if output.Error != "" {
		return nil,
			fmt.Errorf("API responded with error message: %s", output.Error)
	}
	if output.Status != "ok" {
		return nil,
			fmt.Errorf("API responded with non-OK status: %s", output.Status)
	}

	return &output.Data, nil
}
