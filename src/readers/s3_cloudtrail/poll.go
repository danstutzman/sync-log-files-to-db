package s3_cloudtrail

import (
	"strings"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/s3"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
	googleBigquery "google.golang.org/api/bigquery/v2"
)

const SECONDS_BETWEEN_POLLS = 5 * 60

func PollForever(opts *Options, configPath string) {
	s3Conn := s3.NewS3Connection(opts.S3, configPath)

	var bigqueryConn *bigquery.BigqueryConnection
	if opts.BigQuery != nil {
		bigqueryConn = bigquery.NewBigqueryConnection(opts.BigQuery, configPath)
		bigqueryConn.CreateDataset()
		createTable(bigqueryConn)
	}

	var influxdbConn *influxdb.InfluxdbConnection
	if opts.InfluxDb != nil {
		influxdbConn = influxdb.NewInfluxdbConnection(opts.InfluxDb, configPath)
		influxdbConn.CreateDatabase()
	}

	for {
		events := []map[string]interface{}{}

		s3Paths := s3Conn.ListPaths("", int64(opts.PathsPerBatch))
		for _, s3Path := range s3Paths {
			if strings.HasSuffix(s3Path, "/") {
				// Ignore it
			} else if strings.Contains(s3Path, "/CloudTrail/") {
				reader := s3Conn.DownloadPath(s3Path)
				events = append(events, readJsonIntoEventMaps(reader)...)
			} else if strings.Contains(s3Path, "/CloudTrail-Digest/") {
				// ignore it
			} else {
				log.Fatalw("Unexpected S3 path", "path", s3Path)
			}
		}

		if len(events) > 0 {
			if bigqueryConn != nil {
				bigqueryConn.InsertRows(events, "eventID")
			}
			if influxdbConn != nil {
				influxdbConn.InsertMaps(map[string]bool{}, events)
			}
		}

		for _, s3Path := range s3Paths {
			s3Conn.DeletePath(s3Path)
		}

		log.Infow("Wait for next S3 batch...", "seconds", SECONDS_BETWEEN_POLLS)
		time.Sleep(SECONDS_BETWEEN_POLLS * time.Second)
	}
}

func createTable(bigqueryConn *bigquery.BigqueryConnection) {
	bigqueryConn.CreateTable([]*googleBigquery.TableFieldSchema{
		{Name: "timestamp", Type: "TIMESTAMP", Mode: "REQUIRED"},
		{Name: "eventType", Type: "STRING", Mode: "REQUIRED"},
		{Name: "eventName", Type: "STRING", Mode: "REQUIRED"},
		{Name: "sourceIPAddress", Type: "STRING", Mode: "REQUIRED"},
		{Name: "userAgent", Type: "STRING", Mode: "REQUIRED"},
		{Name: "requestId", Type: "STRING", Mode: "REQUIRED"},
		{Name: "eventSource", Type: "STRING", Mode: "REQUIRED"},
		{Name: "eventId", Type: "STRING", Mode: "REQUIRED"},
		{Name: "awsRegion", Type: "STRING", Mode: "REQUIRED"},
		{Name: "responseElements", Type: "STRING", Mode: "REQUIRED"},
		{Name: "requestParameters", Type: "STRING", Mode: "REQUIRED"},
		{Name: "eventVersion", Type: "STRING", Mode: "REQUIRED"},
		{Name: "recipientAccountId", Type: "STRING", Mode: "REQUIRED"},
		{Name: "userIdentityType", Type: "STRING", Mode: "REQUIRED"},
		{Name: "userIdentityAccountId", Type: "STRING", Mode: "REQUIRED"},
		{Name: "userIdentityArn", Type: "STRING", Mode: "REQUIRED"},
		{Name: "userIdentityPrincipalId", Type: "STRING", Mode: "REQUIRED"},
		{Name: "userIdentitySessionContextAttributes", Type: "STRING", Mode: "REQUIRED"},
		{Name: "userIdentityAccessKeyId", Type: "STRING", Mode: "REQUIRED"},
	})
}
