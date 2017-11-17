package s3_cloudtrail

import (
	"log"
	"strings"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

const SECONDS_BETWEEN_POLLS = 5 * 60

func PollForever(opts *Options, configPath string) {
	s3Conn := s3.NewS3Connection(opts.S3, configPath)

	var bigqueryConn *bigquery.BigqueryConnection
	if opts.BigQuery != nil {
		bigqueryConn = bigquery.NewBigqueryConnection(opts.BigQuery, configPath)
		bigqueryConn.CreateDataset()
		bigqueryConn.CreateVisitsTable()
	}

	var influxdbConn *influxdb.InfluxdbConnection
	if opts.InfluxDb != nil {
		influxdbConn = influxdb.NewInfluxdbConnection(opts.InfluxDb, configPath)
		influxdbConn.CreateDatabase()
	}

	for {
		events := []map[string]interface{}{}
		s3Paths := s3Conn.ListPaths("AWSLogs/553826207523/CloudTrail/", int64(opts.PathsPerBatch))
		for _, s3Path := range s3Paths {
			log.Printf(s3Path)
			if strings.HasSuffix(s3Path, "/") {
				// Ignore it
			} else if strings.Contains(s3Path, "/CloudTrail-Digest/") {
				// Ignore it
			} else {
				reader := s3Conn.DownloadPath(s3Path)
				events = append(events, readJsonIntoEventMaps(reader)...)
			}
		}

		if len(events) > 0 {
			if bigqueryConn != nil {
				// bigqueryConn.UploadMaps(events)
			}
			if influxdbConn != nil {
				influxdbConn.InsertMaps(map[string]bool{}, events)
			}
		}
		// for _, s3Path := range s3Paths {
		// 	s3Conn.DeletePath(s3Path)
		// }

		log.Printf("Wait %ds for next S3 batch...", SECONDS_BETWEEN_POLLS)
		time.Sleep(SECONDS_BETWEEN_POLLS * time.Second)
	}
}
