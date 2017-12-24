package redis

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/postgres"
)

type Options struct {
	ListenPort       string
	ExpectedPassword string
	InfluxDb         *influxdb.Options
	Postgresql       *postgres.Options
}

func ValidateOptions(options *Options) {
	if options.ListenPort == "" {
		log.Fatalw("Missing ListenPort")
	}

	if options.InfluxDb != nil {
		influxdb.ValidateOptions(options.InfluxDb)
	}
	if options.Postgresql != nil {
		postgres.ValidateOptions(options.Postgresql)
	}
}
