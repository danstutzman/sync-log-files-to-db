package s3_cloudtrail

import (
	"log"

	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

const DEFAULT_PATHS_PER_BATCH = 100

type Options struct {
	S3            *s3.Options
	BigQuery      *bigquery.Options
	InfluxDb      *influxdb.Options
	PathsPerBatch int
}

func ValidateOptions(options *Options) {
	if options.BigQuery == nil && options.InfluxDb == nil {
		log.Fatalf("Specify either S3.BigQuery or S3.InfluxDb")
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
