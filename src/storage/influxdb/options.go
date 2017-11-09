package influxdb

import (
	"log"
)

type Options struct {
	Hostname        string
	Port            string
	Username        string
	Password        string
	DatabaseName    string
	MeasurementName string
}

func Usage() string {
	return `{
      "Hostname":        STRING,  e.g. "localhost"
      "Port   ":         STRING,  e.g. "8086"
      "Username   ":     STRING,
			"Password":        STRING,
			"DatabaseName":    STRING,
			"MeasurementName": STRING
    }`
}

func ValidateOptions(options *Options) {
	if options.Hostname == "" {
		log.Fatalf("Missing Influxdb.Hostname")
	}
	if options.Port == "" {
		log.Fatalf("Missing Influxdb.Port")
	}
	if options.DatabaseName == "" {
		log.Fatalf("Missing Influxdb.DatabaseName")
	}
}
