package systemd

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
)

type Options struct {
	UnitNames []string
	BigQuery  *bigquery.Options
	InfluxDb  *influxdb.Options
}

func ValidateOptions(options *Options) {
	if len(options.UnitNames) == 0 {
		log.Fatalw("Missing options.UnitNames")
	}
}
