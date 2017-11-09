package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
	my_s3 "github.com/danielstutzman/sync-log-files-to-db/src/storage/s3"
)

const NUM_PATHS_PER_PAGE = 1000

type Config struct {
	BigQuery *bigquery.Options
	S3       *my_s3.Options
	Influxdb *influxdb.Options
}

func readConfig() (*Config, string) {
	if len(os.Args) < 1+1 {
		log.Fatalf("First argument should be config.json")
	}
	configPath := os.Args[1]

	var config = &Config{}
	configJson, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error from ioutil.ReadFile: %s", err)
	}
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		log.Fatalf("Error from json.Unmarshal: %s", err)
	}
	return config, configPath
}

func setupConnections(config *Config, configPath string) (*bigquery.BigqueryConnection,
	*my_s3.S3Connection, *influxdb.InfluxdbConnection) {

	var bigqueryConn *bigquery.BigqueryConnection
	if config.BigQuery != nil {
		bigquery.ValidateOptions(config.BigQuery)
		bigqueryConn = bigquery.NewBigqueryConnection(config.BigQuery, configPath)
	}

	my_s3.ValidateOptions(config.S3)
	s3Connection := my_s3.NewS3Connection(config.S3, configPath)

	influxdb.ValidateOptions(config.Influxdb)
	influxdbConn := influxdb.NewInfluxdbConnection(config.Influxdb, configPath)

	return bigqueryConn, s3Connection, influxdbConn
}

func collectVisits(s3Connection *my_s3.S3Connection) ([]map[string]string, []string) {
	visits := []map[string]string{}
	s3Paths := s3Connection.ListPaths(NUM_PATHS_PER_PAGE)
	for _, s3Path := range s3Paths {
		visits = append(visits, s3Connection.DownloadVisitsForPath(s3Path)...)
	}
	return visits, s3Paths
}

func main() {
	config, configPath := readConfig()

	bigqueryConn, s3Connection, influxdbConn := setupConnections(
		config, configPath)

	visits, s3Paths := collectVisits(s3Connection)

	if len(visits) > 0 {
		if bigqueryConn != nil {
			bigqueryConn.UploadVisits(visits)
		}
		if influxdbConn != nil {
			influxdbConn.InsertVisits(visits)
		}
	}
	for _, s3Path := range s3Paths {
		s3Connection.DeletePath(s3Path)
	}
}
