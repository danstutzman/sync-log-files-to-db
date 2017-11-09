package influxdb

import (
	"log"
	"regexp"
	"strconv"

	clientPkg "github.com/influxdata/influxdb/client/v2"
)

var INTEGER_REGEXP = regexp.MustCompile("^[0-9]+$")
var FLOAT_REGEXP = regexp.MustCompile("^[0-9]+\\.[0-9]+$")

type InfluxdbConnection struct {
	client          clientPkg.Client
	databaseName    string
	measurementName string
}

func NewInfluxdbConnection(opts *Options, configPath string) *InfluxdbConnection {
	client, err := clientPkg.NewHTTPClient(clientPkg.HTTPConfig{
		Addr:     "http://" + opts.Hostname + ":" + opts.Port,
		Username: opts.Username,
		Password: opts.Password,
	})
	if err != nil {
		log.Fatalf("Error from NewHTTPClient: %s", err)
	}

	return &InfluxdbConnection{
		client:          client,
		databaseName:    opts.DatabaseName,
		measurementName: opts.MeasurementName,
	}
}

func (conn *InfluxdbConnection) query(command string) (result []clientPkg.Result, err error) {

	q := clientPkg.Query{
		Command:  command,
		Database: conn.databaseName,
	}
	if response, err := conn.client.Query(q); err == nil {
		if response.Error() != nil {
			return []clientPkg.Result{}, response.Error()
		}
		return response.Results, nil
	}
	return []clientPkg.Result{}, nil
}

func toInt(key, value string) int64 {
	if !INTEGER_REGEXP.MatchString(value) {
		log.Fatalf("Unexpected characters in field %s: '%s'", key, value)
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatalf("Error from Atoi: %s", err)
	}
	return i
}

func toFloat(key, value string) float64 {
	if !FLOAT_REGEXP.MatchString(value) {
		log.Fatalf("Unexpected characters in field %s: '%s'", key, value)
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatalf("Error from ParseFloat: %s", err)
	}
	return f
}
