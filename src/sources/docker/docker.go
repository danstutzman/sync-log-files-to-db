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
var INFLUXDB_TAGS_SET = map[string]bool{
	"container_id": true,
	"image_name":   true,
}
var LOGS_TIMEOUT = time.Duration(1 * time.Second)

func tailContainer(container *types.Container,
	influxdbConn *influxdb.InfluxdbConnection, client *client.Client,
	logLinesChan chan<- LogLine) {

	lastTimestamp := influxdbConn.QueryForLastTimestamp(container.ID)
	justAfterLastTimestamp := lastTimestamp.Add(time.Nanosecond)

	log.Printf("Tailing logs for container %s (image=%s) after %s...",
		container.ID[:10], container.Image, lastTimestamp)

	reader, err := client.ContainerLogs(
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

	noTimeoutChan := make(chan bool, 1)
	go tailLogLines(reader, container.ID, container.Image, noTimeoutChan, logLinesChan)
	select {
	case <-noTimeoutChan:
		// Great, successful read without timeout
	case <-time.After(LOGS_TIMEOUT):
		log.Fatalf("Timeout after %s trying to read logs from container %s (image=%s)",
			LOGS_TIMEOUT, container.ID, container.Image)
	}
}

func pollForNewContainersForever(client *client.Client,
	influxdbConn *influxdb.InfluxdbConnection, logLinesChan chan<- LogLine,
	config *Options) {

	seenContainerIds := map[string]bool{}
	for {
		containers, err := client.ContainerList(
			context.Background(), types.ContainerListOptions{All: true})
		if err != nil {
			log.Fatalf("Error from ContainerList: %s", err)
		}

		for _, container := range containers {
			if !seenContainerIds[container.ID] {
				seenContainerIds[container.ID] = true
				tailContainer(&container, influxdbConn, client, logLinesChan)
			}
		}
		log.Printf("Wait %ds before polling for new containers...",
			config.SecondsBetweenPollForNewContainers)
		time.Sleep(time.Duration(config.SecondsBetweenPollForNewContainers) * time.Second)
	}
}

func syncToInfluxdbForever(logLinesChan <-chan LogLine,
	influxdbConn *influxdb.InfluxdbConnection) {

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
				// No new logs after timeout, so inserting
				influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, maps)
				maps = []map[string]interface{}{}
			}
		}
	}
}

func TailDockerLogsForever(config *Options, configPath string) {
	var influxdbConn *influxdb.InfluxdbConnection
	if config.InfluxDb != nil {
		influxdb.ValidateOptions(config.InfluxDb)
		influxdbConn = influxdb.NewInfluxdbConnection(config.InfluxDb, configPath)
	}
	influxdbConn.CreateDatabase()

	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Error from NewEnvClient: %s", err)
	}

	logLinesChan := make(chan LogLine)
	go pollForNewContainersForever(client, influxdbConn, logLinesChan, config)
	syncToInfluxdbForever(logLinesChan, influxdbConn)
}

func appendLogLineToMaps(logLine LogLine, maps []map[string]interface{}) []map[string]interface{} {
	logLineAsMap := map[string]interface{}{
		"timestamp":    logLine.Timestamp,
		"container_id": logLine.ContainerId,
		"image_name":   logLine.ImageName,
		"message":      logLine.Message,
	}
	return append(maps, logLineAsMap)
}
