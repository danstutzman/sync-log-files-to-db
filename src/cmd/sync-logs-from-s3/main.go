package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/danielstutzman/sync-logs-from-s3/src/storage/bigquery"
	my_s3 "github.com/danielstutzman/sync-logs-from-s3/src/storage/s3"
)

const NUM_PATHS_PER_PAGE = 100

func main() {
	bigqueryOptionsBytes, err := ioutil.ReadFile("config/bigquery.json")
	if err != nil {
		log.Fatalf("Error from ReadFile: %s", err)
	}
	var bigqueryOptions bigquery.Options
	json.Unmarshal(bigqueryOptionsBytes, &bigqueryOptions)
	bigquery.ValidateOptions(&bigqueryOptions)
	bigqueryConn := bigquery.NewBigqueryConnection(&bigqueryOptions)

	s3OptionsBytes, err := ioutil.ReadFile("config/s3.json")
	if err != nil {
		log.Fatalf("Error from ReadFile: %s", err)
	}
	var s3Options my_s3.Options
	json.Unmarshal(s3OptionsBytes, &s3Options)
	my_s3.ValidateOptions(&s3Options)
	s3Connection := my_s3.NewS3Connection(&s3Options)

	visits := []map[string]string{}
	pageOfPaths := s3Connection.ListPaths(NUM_PATHS_PER_PAGE)
	for _, path := range pageOfPaths {
		visits = append(visits, s3Connection.DownloadVisitsForPath(path)...)
	}
	if len(visits) > 0 {
		bigqueryConn.UploadVisits(visits)
	}
	for _, path := range pageOfPaths {
		_ = path
		// s3Connection.DeletePath(path)
	}
}
