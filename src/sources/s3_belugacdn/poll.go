package s3_belugacdn

import (
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
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
		createVisitsTable(bigqueryConn)
	}

	var influxdbConn *influxdb.InfluxdbConnection
	if opts.InfluxDb != nil {
		influxdbConn = influxdb.NewInfluxdbConnection(opts.InfluxDb, configPath)
		influxdbConn.CreateDatabase()
	}

	for {
		visits := []map[string]interface{}{}
		s3Paths := s3Conn.ListPaths("", int64(opts.PathsPerBatch))
		for _, s3Path := range s3Paths {
			reader := s3Conn.DownloadPath(s3Path)
			visits = append(visits, readJsonIntoVisitMaps(reader)...)
		}

		if len(visits) > 0 {
			if bigqueryConn != nil {
				bigqueryConn.InsertRows("test", visits, "trace")
			}
			if influxdbConn != nil {
				influxdbConn.InsertMaps(VISITS_TAG_SET, visits)
			}
		}
		for _, s3Path := range s3Paths {
			s3Conn.DeletePath(s3Path)
		}

		if opts.RunOnce {
			break
		}

		log.Infow("Wait for next S3 batch...", "seconds", SECONDS_BETWEEN_POLLS)
		time.Sleep(SECONDS_BETWEEN_POLLS * time.Second)
	}
}
