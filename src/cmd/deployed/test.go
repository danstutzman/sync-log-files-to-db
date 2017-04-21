package main

import (
	"encoding/json"
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/src/storage/bigquery"
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/src/storage/s3"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 1+1 {
		log.Fatalf("Expected arg 1 to be JSON of event")
	}
	events := decodeEvents(os.Args[1])

	optionsBytes, err := ioutil.ReadFile("config/bigquery.json")
	if err != nil {
		log.Fatalf("Error from ReadFile: %s", err)
	}
	var options bigquery.Options
	json.Unmarshal(optionsBytes, &options)
	bigquery.ValidateOptions(&options)
	bigqueryConn := bigquery.NewBigqueryConnection(&options)
	log.Printf("Results from SELECT 1: %v",
		bigqueryConn.Query("SELECT 1", "SELECT 1"))

	log.Printf("HELLO FROM GOLANG WITH ARGS %v\n", os.Args)

	for _, record := range events.Records {
		if record.EventSource == "aws:s3" &&
			record.EventName == "ObjectCreated:Put" {
			s3Connection := s3.NewS3Connection(record.S3.Bucket.Name)
			//log.Printf("ListPaths: %v", s3Connection.ListPaths())
			visits := downloadVisitsForPath(s3Connection, record.S3.Object.Key)
			log.Printf("Visits: %s", visits)

		} else {
			log.Fatalf("Don't know how to handle event EventSource:%s EventName:%s",
				record.EventSource, record.EventName)
		}
	}
}
