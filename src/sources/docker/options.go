package docker

import (
	"log"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

const DEFAULT_SECONDS_BETWEEN_POLL_FOR_NEW_CONTAINERS = 60

type Options struct {
	InfluxDb                           *influxdb.Options
	SecondsBetweenPollForNewContainers int
}

func ValidateOptions(options *Options) {
	if options.SecondsBetweenPollForNewContainers == 0 {
		options.SecondsBetweenPollForNewContainers =
			DEFAULT_SECONDS_BETWEEN_POLL_FOR_NEW_CONTAINERS
	}

	if options.InfluxDb == nil {
		log.Fatalf("Missing Docker.InfluxDb")
	}
	influxdb.ValidateOptions(options.InfluxDb)
}
