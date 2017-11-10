package docker

import (
	"context"
	"log"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
)

const MAX_INFLUXDB_INSERT_BATCH_SIZE = 100

var TAIL_LOG_LINE_FLUSH_TIMEOUT = time.Millisecond * 100
var INFLUXDB_TAGS_SET = map[string]bool{"image_name": true}

func TailDockerLogs(config *Options, configPath string) {
	var influxdbConn *influxdb.InfluxdbConnection
	if config.Influxdb != nil {
		influxdb.ValidateOptions(config.Influxdb)
		influxdbConn = influxdb.NewInfluxdbConnection(config.Influxdb, configPath)
	}
	influxdbConn.CreateDatabase()

	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Error from NewEnvClient: %s", err)
	}

	containers, err := client.ContainerList(
		context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Error from ContainerList: %s", err)
	}

	logLinesChan := make(chan LogLine)
	for _, container := range containers {
		lastTimestamp := influxdbConn.QueryForLastTimestamp(container.Image)
		justAfterLastTimestamp := lastTimestamp.Add(time.Nanosecond)

		log.Printf("Tailing logs for container %s (image=%s) after %s...",
			container.ID[:10], container.Image, lastTimestamp)

		out, err := client.ContainerLogs(
			context.Background(),
			container.ID,
			types.ContainerLogsOptions{
				Details:    true,
				Follow:     true,
				ShowStdout: true,
				ShowStderr: true,
				Since:      justAfterLastTimestamp.Format(time.RFC3339Nano),
				Timestamps: true,
			})
		if err != nil {
			panic(err)
		}

		go tailLogLines(out, container.Image, logLinesChan)
	}

	maps := []map[string]interface{}{}
	for {
		if len(maps) == 0 {
			logLine := <-logLinesChan
			maps = appendLogLineToMaps(logLine, maps)
		} else if len(maps) >= MAX_INFLUXDB_INSERT_BATCH_SIZE {
			influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, maps)
			maps = []map[string]interface{}{}
		} else {
			select {
			case logLine := <-logLinesChan:
				maps = appendLogLineToMaps(logLine, maps)
			case <-time.After(TAIL_LOG_LINE_FLUSH_TIMEOUT):
				log.Printf("No new logs after %v timeout, so inserting",
					TAIL_LOG_LINE_FLUSH_TIMEOUT)
				influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, maps)
				maps = []map[string]interface{}{}
			}
		}
	}
}

func appendLogLineToMaps(logLine LogLine, maps []map[string]interface{}) []map[string]interface{} {
	logLineAsMap := map[string]interface{}{
		"timestamp":  logLine.Timestamp,
		"image_name": logLine.ImageName,
		"message":    logLine.Message,
	}
	return append(maps, logLineAsMap)
}
