package redis

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

type Options struct {
	ListenPort       string
	ExpectedPassword string
	InfluxDb         *influxdb.Options
}

func ValidateOptions(options *Options) {
	if options.ListenPort == "" {
		log.Fatalw("Missing ListenPort")
	}

	if options.InfluxDb != nil {
		influxdb.ValidateOptions(options.InfluxDb)
	}
}
