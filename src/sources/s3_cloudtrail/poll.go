package s3_cloudtrail

import (
	"log"
	"strings"

	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3"
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
		createEventsTable(bigqueryConn)
		createDigestsTable(bigqueryConn)
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
				log.Fatalf("Unexpected path %s", s3Path)
			}
		}

		if len(events) > 0 {
			if bigqueryConn != nil {
				bigqueryConn.InsertRows("cloudtrail_events4",
					func() { bigqueryConn.CreateDataset() },
					func() { createEventsTable(bigqueryConn) },
					events,
					"eventID")
			}
			if influxdbConn != nil {
				influxdbConn.InsertMaps(map[string]bool{}, events)
			}
		}

		for _, s3Path := range s3Paths {
			s3Conn.DeletePath(s3Path)
		}

		// log.Printf("Wait %ds for next S3 batch...", SECONDS_BETWEEN_POLLS)
		// time.Sleep(SECONDS_BETWEEN_POLLS * time.Second)
	}
}

func createEventsTable(bigqueryConn *bigquery.BigqueryConnection) {
	bigqueryConn.CreateTable("cloudtrail_events4", []*googleBigquery.TableFieldSchema{
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

func createDigestsTable(bigqueryConn *bigquery.BigqueryConnection) {
	bigqueryConn.CreateTable("cloudtrail_digests", []*googleBigquery.TableFieldSchema{
		{Name: "timestamp", Type: "TIMESTAMP", Mode: "REQUIRED"},
		{Name: "awsAccountId", Type: "STRING", Mode: "REQUIRED"},
		{Name: "digestStartTime", Type: "STRING", Mode: "REQUIRED"},
		{Name: "digestEndTime", Type: "STRING", Mode: "REQUIRED"},
		{Name: "digestS3Bucket", Type: "STRING", Mode: "REQUIRED"},
		{Name: "digestS3Object", Type: "STRING", Mode: "REQUIRED"},
		{Name: "digestPublicKeyFingerprint", Type: "STRING", Mode: "REQUIRED"},
		{Name: "digestSignatureAlgorithm", Type: "STRING", Mode: "REQUIRED"},
		{Name: "newestEventTime", Type: "STRING", Mode: "REQUIRED"},
		{Name: "oldestEventTime", Type: "STRING", Mode: "REQUIRED"},
		{Name: "previousDigestS3Bucket", Type: "STRING", Mode: "REQUIRED"},
		{Name: "previousDigestS3Object", Type: "STRING", Mode: "REQUIRED"},
		{Name: "previousDigestHashValue", Type: "STRING", Mode: "REQUIRED"},
		{Name: "previousDigestHashAlgorithm", Type: "STRING", Mode: "REQUIRED"},
		{Name: "previousDigestSignature", Type: "STRING", Mode: "REQUIRED"},
		{Name: "LogFiles", Type: "STRING", Mode: "REQUIRED"},
	})
}
