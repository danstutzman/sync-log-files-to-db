package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/danielstutzman/sync-logs-from-s3/src/storage/bigquery"
	my_s3 "github.com/danielstutzman/sync-logs-from-s3/src/storage/s3"
)

const NUM_PATHS_PER_PAGE = 1000

type Config struct {
	Bigquery bigquery.Options
	S3       my_s3.Options
}

func main() {
	if len(os.Args) < 1+1 {
		log.Fatalf("First argument should be config.json")
	}

	var config = &Config{}
	configJson, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error from ioutil.ReadFile: %s", err)
	}
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		log.Fatalf("Error from json.Unmarshal: %s", err)
	}

	bigquery.ValidateOptions(&config.Bigquery)
	bigqueryConn := bigquery.NewBigqueryConnection(&config.Bigquery)

	my_s3.ValidateOptions(&config.S3)
	s3Connection := my_s3.NewS3Connection(&config.S3)

	visits := []map[string]string{}
	pageOfPaths := s3Connection.ListPaths(NUM_PATHS_PER_PAGE)
	for _, path := range pageOfPaths {
		visits = append(visits, s3Connection.DownloadVisitsForPath(path)...)
	}
	if len(visits) > 0 {
		bigqueryConn.UploadVisits(visits)
	}
	for _, path := range pageOfPaths {
		s3Connection.DeletePath(path)
	}
}
