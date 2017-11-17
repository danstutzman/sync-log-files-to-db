package s3_belugacdn

import (
	"compress/gzip"
	"fmt"
	"log"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

func PollForever(opts *Options, configPath string) {
	s3Conn := s3.NewS3Connection(opts.S3, configPath)

	var bigqueryConn *bigquery.BigqueryConnection
	if opts.BigQuery != nil {
		bigqueryConn = bigquery.NewBigqueryConnection(opts.BigQuery, configPath)
	}

	var influxdbConn *influxdb.InfluxdbConnection
	if opts.InfluxDb != nil {
		influxdbConn = influxdb.NewInfluxdbConnection(opts.InfluxDb, configPath)
		influxdbConn.CreateDatabase()
	}

	for {
		visits := []map[string]string{}
		s3Paths := s3Conn.ListPaths("", int64(opts.PathsPerBatch))
		for _, s3Path := range s3Paths {
			reader := s3Conn.DownloadPath(s3Path)
			visits = append(visits, readJsonIntoVisitMaps(reader)...)
		}

		if len(visits) > 0 {
			if bigqueryConn != nil {
				bigqueryConn.UploadVisits(visits)
			}
			if influxdbConn != nil {
				influxdbConn.InsertVisits(visits)
			}
		}
		for _, s3Path := range s3Paths {
			s3Conn.DeletePath(s3Path)
		}

		log.Printf("Wait %ds for next S3 batch...", opts.SecondsBetweenPolls)
		time.Sleep(time.Duration(opts.SecondsBetweenPolls) * time.Second)
	}
}
