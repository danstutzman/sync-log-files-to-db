package redis

import (
	"log"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

type Options struct {
	ListenPort       string
	ExpectedPassword string
	InfluxDb         *influxdb.Options
}

func ValidateOptions(options *Options) {
	if options.ListenPort == "" {
		log.Fatalf("Missing ListenPort")
	}

	if options.InfluxDb != nil {
		influxdb.ValidateOptions(options.InfluxDb)
	}
}
