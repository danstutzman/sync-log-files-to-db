package influxdb

import (
	"fmt"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
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
		log.Fatalw("Error from NewHTTPClient", "err", err)
	}

	return &InfluxdbConnection{
		client:          client,
		databaseName:    opts.DatabaseName,
		measurementName: opts.MeasurementName,
	}
}

func (conn *InfluxdbConnection) CreateDatabase() {
	log.Infow("Creating InfluxDB database...", "databaseName", conn.databaseName)
	command := fmt.Sprintf("CREATE DATABASE %s", conn.databaseName)
	_, err := conn.query(command)
	if err != nil {
		log.Fatalw("Error from command", "command", command, "err", err)
	}
}
