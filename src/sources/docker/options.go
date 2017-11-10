package docker

import (
	"log"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

type Options struct {
	InfluxDb *influxdb.Options
}

func ValidateOptions(options *Options) {
	if options.InfluxDb == nil {
		log.Fatalf("Missing Docker.InfluxDb")
	}
	influxdb.ValidateOptions(options.InfluxDb)
}
