[![GoDoc](https://godoc.org/github.com/danielstutzman/go-monitis?status.svg)](https://godoc.org/github.com/danielstutzman/go-monitis)

# Go SDK for Monitis API

## Supported
* Authentication:
  * Get User AuthToken [(API docs)](http://www.monitis.com/docs/apiActions.html#getAuthToken)
* Predefined Monitors:
  * External Monitors:
    * Add External Monitor [(API docs)](http://www.monitis.com/docs/apiActions.html#addExternalMonitor)
    * Edit External Monitor [(API docs)](http://www.monitis.com/docs/apiActions.html#editExternalMonitor)
    * Delete External Monitors [(API docs)](http://www.monitis.com/docs/apiActions.html#deleteExternalMonitor)
    * Get Locations [(API docs)](http://www.monitis.com/docs/apiActions.html#getExternalMonitorLocations)
    * Get List of External Monitors [(API docs)](http://www.monitis.com/docs/apiActions.html#getExternalMonitors)
    * Get External Monitor Info [(API docs)](http://www.monitis.com/docs/apiActions.html#getExternalMonitorInfo)
    * Get External Monitor Results [(API docs)](http://www.monitis.com/docs/apiActions.html#getExternalMonitorResults)
* Contacts API:
  * Get Recent Alerts [(API docs)](http://www.monitis.com/docs/apiActions.html#getRecentAlerts)

## Installation
```
go get -u github.com/danielstutzman/go-monitis
```

## Command-line interface for testing
```
$GOPATH/bin/monitis-cli -apikey API_KEY_HERE -secretkey SECRET_KEY_HERE
```

## See also
* [Monitis REST API documentation](http://www.monitis.com/docs/api.html)
