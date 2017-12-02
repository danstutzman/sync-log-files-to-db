package docker

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"regexp"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

const DOCKER_LOG_TIME_FORMAT = "2006-01-02T15:04:05.999999999Z"

var STREAM_TYPE_DESCRIPTIONS = []string{"STDIN", "STDOUT", "STDERR"}

var DOCKER_LOG_LINE_REGEXP = regexp.MustCompile(
	"^([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\\.[0-9]{9}Z)" +
		"  (.*)\n$")

type LogLine struct {
	ContainerId string
	ImageName   string
	StreamType  string
	Timestamp   time.Time
	Message     string
}

type JsonFileLog struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
	Time   string `json:"time"`
}

func tailLogLinesForJsonFile(out io.Reader, containerId, imageName string,
	logLinesChan chan<- LogLine) {

	log.Infow("tailLogLinesForJsonFile", "container_id", containerId[:10], "image_name", imageName)

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalw("Error from scanner.Err", "err", scanner.Err())
		}
		jsonFileLogJson := scanner.Bytes()

		jsonFileLog := JsonFileLog{}
		err := json.Unmarshal(jsonFileLogJson, &jsonFileLog)
		if err != nil {
			log.Fatalw("Error from json.Unmarshal", "err", err)
		}

		timestamp, err := time.Parse(DOCKER_LOG_TIME_FORMAT, jsonFileLog.Time)
		if err != nil {
			log.Fatalw("Can't parse timestamp", "time", jsonFileLog.Time)
		}

		logLinesChan <- LogLine{
			StreamType:  jsonFileLog.Stream,
			Timestamp:   timestamp,
			Message:     jsonFileLog.Log,
			ContainerId: containerId,
			ImageName:   imageName,
		}
	}
}

func tailLogLines(out io.Reader, containerId, imageName string,
	noTimeoutChan chan<- bool, logLinesChan chan<- LogLine) {
	log.Infow("tailLogLines", "container_id", containerId[:10], "image_name", imageName)

	reader := bufio.NewReader(out)
	possibleHeader, err := reader.Peek(8)
	noTimeoutChan <- true
	if err != nil {
		if err == io.EOF {
			log.Infow("EOF from logs", "containerId", containerId[:10])
			return
		} else {
			log.Fatalw("Error from Peek", "err", err)
		}
	}

	if possibleHeader[0] > 2 ||
		possibleHeader[1] != 0 ||
		possibleHeader[2] != 0 ||
		possibleHeader[3] != 0 {

		for {
			// not using headers
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatalw("Error from ReadBytes", "err", err)
			}

			// Parse timestamp in line
			match := DOCKER_LOG_LINE_REGEXP.FindSubmatch(line)
			timestamp, err := time.Parse(DOCKER_LOG_TIME_FORMAT, string(match[1]))
			message := string(match[2])
			if err != nil {
				log.Fatalw("Can't parse timestamp", "timestamp", match[1])
			}

			logLine := LogLine{
				ContainerId: containerId,
				ImageName:   imageName,
				Message:     message,
				Timestamp:   timestamp,
			}
			logLinesChan <- logLine
		}
	} else {
		// using headers
		for {
			logLine := readLogLineBlocking(reader)
			if logLine == nil {
				break
			} else {
				logLine.ContainerId = containerId
				logLine.ImageName = imageName
				logLinesChan <- *logLine
			}
		}
	}
}

func readLogLineBlocking(out io.Reader) *LogLine {
	// Read 8-byte header (blocking)
	// See https://docs.docker.com/engine/api/v1.24/ for header format
	header := make([]byte, 8)
	numBytes, err := io.ReadFull(out, header)
	//numBytes, err := out.Read(header)
	if numBytes == 0 {
		return nil
	}
	if numBytes < 8 {
		log.Fatalw("Expected 8 but got", "num_bytes", numBytes)
	}
	if err != nil {
		log.Fatalw("Error from Read", "err", err)
	}

	// Parse files out of header
	streamType := STREAM_TYPE_DESCRIPTIONS[header[0]]
	lineLen := binary.BigEndian.Uint32(header[4:8])

	// Read line (blocking)
	line := make([]byte, lineLen)
	//numBytes, err = out.Read(line)
	numBytes, err = io.ReadFull(out, line)
	if uint32(numBytes) < lineLen {
		line = append(line[0:numBytes], []byte("(truncated)")...)
	}
	if err != nil {
		log.Fatalw("Error from Read", "err", err)
	}

	// Parse timestamp in line
	match := DOCKER_LOG_LINE_REGEXP.FindSubmatch(line)
	if len(match) == 0 {
		log.Fatalw("No DOCKER_LOG_LINE_REGEXP match", "line", line)
		return nil
	} else {
		timestamp, err := time.Parse(DOCKER_LOG_TIME_FORMAT, string(match[1]))
		message := string(match[2])
		if err != nil {
			log.Fatalw("Can't parse timestamp", "timestamp", match[1])
		}

		return &LogLine{
			StreamType: streamType,
			Timestamp:  timestamp,
			Message:    message,
		}
	}
}
