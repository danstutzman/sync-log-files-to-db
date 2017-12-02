package influxdb

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
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
		log.Fatalw("Missing Influxdb.Hostname")
	}
	if options.Port == "" {
		log.Fatalw("Missing Influxdb.Port")
	}
	if options.DatabaseName == "" {
		log.Fatalw("Missing Influxdb.DatabaseName")
	}
}
