package bigquery

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

type Options struct {
	GcloudPemPath   string
	GcloudProjectId string
	DatasetName     string
	Endpoint        string
}

func ValidateOptions(options *Options) {
	if options.GcloudPemPath == "" {
		log.Fatalw("Missing Bigquery.GcloudPemPath")
	}
	if options.GcloudProjectId == "" {
		log.Fatalw("Missing Bigquery.GcloudProjectId")
	}
	if options.DatasetName == "" {
		log.Fatalw("Missing Bigquery.DatasetName")
	}
}
