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

	for _, container := range containers {
		log.Printf("Tailing logs for container %s (image=%s)...",
			container.ID[:10], container.Image)

		out, err := client.ContainerLogs(
			context.Background(),
			container.ID,
			types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				// Since      string
				Timestamps: true,
				Follow:     true,
				// Tail:       string, // num lines from end to show?
				Details: true,
			})
		if err != nil {
			panic(err)
		}

		logLinesChan := tailLogLines(out)

		maps := []map[string]interface{}{}
		for {
			if len(maps) == 0 {
				logLine := <-logLinesChan
				logLineAsMap := map[string]interface{}{
					"timestamp": logLine.Timestamp,
					"message":   logLine.Message,
				}
				maps = append(maps, logLineAsMap)
			} else if len(maps) >= MAX_INFLUXDB_INSERT_BATCH_SIZE {
				influxdbConn.InsertMaps(maps)
				maps = []map[string]interface{}{}
			} else {
				select {
				case logLine := <-logLinesChan:
					logLineAsMap := map[string]interface{}{
						"timestamp": logLine.Timestamp,
						"message":   logLine.Message,
					}
					maps = append(maps, logLineAsMap)
				case <-time.After(TAIL_LOG_LINE_FLUSH_TIMEOUT):
					log.Println("No new logs after timeout, so inserting")
					influxdbConn.InsertMaps(maps)
					maps = []map[string]interface{}{}
				}
			}
		}
	}
}
