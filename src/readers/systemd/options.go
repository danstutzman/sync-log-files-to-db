package systemd

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/postgres"
)

type Options struct {
	UnitNames  []string
	BigQuery   *bigquery.Options
	InfluxDb   *influxdb.Options
	Postgresql *postgres.Options
}

func ValidateOptions(options *Options) {
	if len(options.UnitNames) == 0 {
		log.Fatalw("Missing TailSystemdLogs.UnitNames")
	}

	if options.InfluxDb == nil &&
		options.Postgresql == nil {
		log.Fatalw("Specify TailSystemdLogs.InfluxDb or TailSystemdLogs.Postgresql")
	}
	if options.InfluxDb != nil {
		influxdb.ValidateOptions(options.InfluxDb)
	}
	if options.Postgresql != nil {
		postgres.ValidateOptions(options.Postgresql)
	}
}
