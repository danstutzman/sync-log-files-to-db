package docker

import (
	"log"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

type Options struct {
	Influxdb *influxdb.Options
}

func ValidateOptions(options *Options) {
	if options.Influxdb == nil {
		log.Fatalf("Missing Docker.Influxdb")
	}
}
