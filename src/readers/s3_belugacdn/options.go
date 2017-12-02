package s3_belugacdn

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/s3"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
)

const DEFAULT_PATHS_PER_BATCH = 100

type Options struct {
	S3            *s3.Options
	BigQuery      *bigquery.Options
	InfluxDb      *influxdb.Options
	PathsPerBatch int
	RunOnce       bool
}

func ValidateOptions(options *Options) {
	if options.BigQuery == nil && options.InfluxDb == nil {
		log.Fatalw("Specify either S3.BigQuery or S3.InfluxDb")
	}
	if options.BigQuery != nil {
		bigquery.ValidateOptions(options.BigQuery)
	}
	if options.InfluxDb != nil {
		influxdb.ValidateOptions(options.InfluxDb)
	}

	if options.PathsPerBatch == 0 {
		options.PathsPerBatch = DEFAULT_PATHS_PER_BATCH
	}
}
