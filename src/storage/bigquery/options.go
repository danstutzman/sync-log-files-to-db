package bigquery

import (
	"log"
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
		log.Fatalf("Missing Bigquery.GcloudPemPath")
	}
	if options.GcloudProjectId == "" {
		log.Fatalf("Missing Bigquery.GcloudProjectId")
	}
	if options.DatasetName == "" {
		log.Fatalf("Missing Bigquery.DatasetName")
	}
}
