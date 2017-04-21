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

	var bigqueryOptions bigquery.Options
	json.Unmarshal(optionsBytes, &bigqueryOptions)
	bigquery.ValidateOptions(&bigqueryOptions)
	bigqueryConn := bigquery.NewBigqueryConnection(&bigqueryOptions)
	log.Printf("Results from SELECT 1: %v",
		bigqueryConn.Query("SELECT 1", "SELECT 1"))

	log.Printf("HELLO FROM GOLANG WITH ARGS %v\n", os.Args)

	for _, record := range events.Records {
		if record.EventSource == "aws:s3" &&
			record.EventName == "ObjectCreated:Put" {
			s3Connection := s3.NewS3Connection(&s3.Options{
				BucketName: record.S3.Bucket.Name,
			})
			//log.Printf("ListPaths: %v", s3Connection.ListPaths())
			s3Path := record.S3.Object.Key
			visits := s3Connection.DownloadVisitsForPath(s3Path)
			bigqueryConn.UploadVisits(s3Path, visits)
			s3Connection.DeletePath(s3Path)
		} else {
			log.Fatalf("Don't know how to handle event EventSource:%s EventName:%s",
				record.EventSource, record.EventName)
		}
	}
}
