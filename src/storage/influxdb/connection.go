package influxdb

import (
	"fmt"
	"log"

	clientPkg "github.com/influxdata/influxdb/client/v2"
)

type InfluxdbConnection struct {
	client          clientPkg.Client
	databaseName    string
	measurementName string
}

func NewInfluxdbConnection(opts *Options, configPath string) *InfluxdbConnection {
	client, err := clientPkg.NewHTTPClient(clientPkg.HTTPConfig{
		Addr:     "http://" + opts.Hostname + ":" + opts.Port,
		Username: opts.Username,
		Password: opts.Password,
	})
	if err != nil {
		log.Fatalf("Error from NewHTTPClient: %s", err)
	}

	return &InfluxdbConnection{
		client:          client,
		databaseName:    opts.DatabaseName,
		measurementName: opts.MeasurementName,
	}
}

func (conn *InfluxdbConnection) CreateDatabase() {
	log.Printf("Creating InfluxDB database %s...", conn.databaseName)
	command := fmt.Sprintf("CREATE DATABASE %s", conn.databaseName)
	_, err := conn.query(command)
	if err != nil {
		log.Fatalf("Error from %s: %s", command, err)
	}
}
