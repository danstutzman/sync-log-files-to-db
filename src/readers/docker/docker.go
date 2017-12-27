package docker

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
)

const MAX_INFLUXDB_INSERT_BATCH_SIZE = 20000

var TAIL_LOG_LINE_FLUSH_TIMEOUT = time.Second
var INFLUXDB_TAGS_SET = map[string]bool{
	"image_name":  true,
	"status_code": true,
}
var LOGS_TIMEOUT = time.Duration(1 * time.Second)

func tailContainer(container *types.Container,
	influxdbConn *influxdb.Connection, client *client.Client,
	logLinesChan chan<- LogLine) {

	lastTimestamp := influxdbConn.QueryForLastTimestampForTag(
		"message", "image_name", container.Image)
	justAfterLastTimestamp := lastTimestamp.Add(time.Nanosecond)

	log.Infow("Tailing logs...", "container_id", container.ID[:10],
		"image_name", container.Image, "after", lastTimestamp)

	inspect, err := client.ContainerInspect(context.TODO(), container.ID)
	if err != nil {
		log.Fatalw("Error from ContainerInspect", "container_id", container.ID[:10], "err", err)
	}

	var reader io.Reader
	log.Infow("Attempting to open log", "path", inspect.LogPath)
	reader, err = os.Open(inspect.LogPath)
	if err == nil {
		tailLogLinesForJsonFile(reader, container.ID, container.Image, logLinesChan)
	} else if os.IsNotExist(err) {
		log.Infow("Open failed (normal if Docker for Mac)", "err", err)
		// reassign reader
		reader, err = client.ContainerLogs(
			context.TODO(),
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
			log.Fatalw("Error from ContainerLogsOptions", "err", err)
		}

		noTimeoutChan := make(chan bool, 1)
		go tailLogLines(reader, container.ID, container.Image, noTimeoutChan, logLinesChan)
		select {
		case <-noTimeoutChan:
			// Great, successful read without timeout
		case <-time.After(LOGS_TIMEOUT):
			log.Fatalw("Timeout trying to read logs", "container_id", container.ID[:10],
				"seconds", LOGS_TIMEOUT, "image_name", container.Image)
		}
	} else {
		log.Fatalw("Error from Open", "err", err)
	}
}

func pollForNewContainersForever(client *client.Client,
	influxdbConn *influxdb.Connection, logLinesChan chan<- LogLine,
	config *Options) {

	seenContainerIds := map[string]bool{}
	for {
		containers, err := client.ContainerList(
			context.TODO(), types.ContainerListOptions{All: true})
		if err != nil {
			log.Fatalw("Error from ContainerList", "err", err)
		}

		for _, container := range containers {
			if !seenContainerIds[container.ID] {
				seenContainerIds[container.ID] = true
				tailContainer(&container, influxdbConn, client, logLinesChan)
			}
		}
		log.Infow("Wait before polling for new containers...",
			"seconds", config.SecondsBetweenPollForNewContainers)
		time.Sleep(time.Duration(config.SecondsBetweenPollForNewContainers) * time.Second)
	}
}

func syncToInfluxdbForever(logLinesChan <-chan LogLine,
	influxdbConn *influxdb.Connection) {

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
	var influxdbConn *influxdb.Connection
	if config.InfluxDb != nil {
		influxdb.ValidateOptions(config.InfluxDb)
		influxdbConn = influxdb.NewConnection(config.InfluxDb, configPath)
	}
	influxdbConn.CreateDatabase()

	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalw("Error from NewEnvClient", "err", err)
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
	}

	augmentMapWithParsedMessage(logLineAsMap, logLine.Message)

	return append(maps, logLineAsMap)
}
