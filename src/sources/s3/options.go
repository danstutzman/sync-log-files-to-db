package s3

import (
	"log"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

const DEFAULT_SECONDS_BETWEEN_POLLS = 5 * 60

type Options struct {
	CredsPath           string
	Region              string
	BucketName          string
	BigQuery            *bigquery.Options
	InfluxDb            *influxdb.Options
	SecondsBetweenPolls int
}

func ValidateOptions(options *Options) {
	if options.CredsPath == "" {
		log.Fatalf("Missing S3.CredsPath")
	}
	if options.Region == "" {
		log.Fatalf("Missing S3.Region")
	}
	if options.BucketName == "" {
		log.Fatalf("Missing S3.BucketName")
	}
	if options.SecondsBetweenPolls == 0 {
		options.SecondsBetweenPolls = DEFAULT_SECONDS_BETWEEN_POLLS
	}

	if options.BigQuery == nil && options.InfluxDb == nil {
		log.Fatalf("Specify either S3.BigQuery or S3.InfluxDb")
	}
	if options.BigQuery != nil {
		bigquery.ValidateOptions(options.BigQuery)
	}
	if options.InfluxDb != nil {
		influxdb.ValidateOptions(options.InfluxDb)
	}
}
