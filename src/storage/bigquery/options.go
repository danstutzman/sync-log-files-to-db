package bigquery

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

type Options struct {
	GcloudPemPath   string
	GcloudProjectId string
	DatasetName     string
}

func Usage() string {
	return `{
      "GcloudPemPath":     STRING,  path to Google JSON creds for BigQuery,
                                      e.g. "./Speech-ba6281533dc8.json"
      "GcloudProjectId":   STRING   Project number or project ID for BigQuery
      "DatasetName":       STRING   Name of BigQuery dataset
    }`
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
