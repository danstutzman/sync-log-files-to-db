package monitis

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/postgres"
)

type Options struct {
	ApiKey     string
	SecretKey  string
	InfluxDb   *influxdb.Options
	Postgresql *postgres.Options
}

func ValidateOptions(options *Options) {
	if options.ApiKey == "" {
		log.Fatalw("Missing Monitis.ApiKey")
	}
	if options.SecretKey == "" {
		log.Fatalw("Missing Monitis.SecretKey")
	}

	if options.InfluxDb == nil &&
		options.Postgresql == nil {
		log.Fatalw("Specify Monitis.InfluxDb and/or Monitis.Postgresql")
	}
	if options.InfluxDb != nil {
		influxdb.ValidateOptions(options.InfluxDb)
	}
	if options.Postgresql != nil {
		postgres.ValidateOptions(options.Postgresql)
	}
}
