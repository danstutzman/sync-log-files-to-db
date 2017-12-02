package monitis

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
)

type Options struct {
	ApiKey    string
	SecretKey string
	InfluxDb  *influxdb.Options
}

func ValidateOptions(options *Options) {
	if options.ApiKey == "" {
		log.Fatalw("Missing Monitis.ApiKey")
	}
	if options.SecretKey == "" {
		log.Fatalw("Missing Monitis.SecretKey")
	}

	influxdb.ValidateOptions(options.InfluxDb)
}
