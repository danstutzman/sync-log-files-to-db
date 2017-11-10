package docker

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"regexp"
	"time"
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

func tailLogLines(out io.Reader, containerId, imageName string, logLinesChan chan<- LogLine) {
	log.Printf("Tail for container %s image %s", containerId, imageName)

	reader := bufio.NewReader(out)
	possibleHeader, err := reader.Peek(8)
	if err != nil {
		if err == io.EOF {
			log.Printf("EOF from logs of %s", containerId)
			return
		} else {
			log.Fatalf("Error from Peek: %s", err)
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
				log.Fatalf("Error from ReadBytes: %s", err)
			}

			// Parse timestamp in line
			match := DOCKER_LOG_LINE_REGEXP.FindSubmatch(line)
			timestamp, err := time.Parse(DOCKER_LOG_TIME_FORMAT, string(match[1]))
			message := string(match[2])
			if err != nil {
				log.Fatalf("Can't parse timestamp %s", string(match[1]))
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
	numBytes, err := out.Read(header)
	if numBytes == 0 {
		return nil
	}
	if numBytes < 8 {
		log.Fatalf("Expected 8 but got %d", numBytes)
	}
	if err != nil {
		log.Fatalf("Error from Read: %s", err)
	}

	// Parse files out of header
	streamType := STREAM_TYPE_DESCRIPTIONS[header[0]]
	lineLen := binary.BigEndian.Uint32(header[4:8])

	// Read line (blocking)
	line := make([]byte, lineLen)
	numBytes, err = out.Read(line)
	if uint32(numBytes) < lineLen {
		log.Fatalf("Expected %d but got %d", lineLen, numBytes)
	}
	if err != nil {
		log.Fatalf("Error from Read: %s", err)
	}

	// Parse timestamp in line
	match := DOCKER_LOG_LINE_REGEXP.FindSubmatch(line)
	timestamp, err := time.Parse(DOCKER_LOG_TIME_FORMAT, string(match[1]))
	message := string(match[2])
	if err != nil {
		log.Fatalf("Can't parse timestamp %s", string(match[1]))
	}

	return &LogLine{
		StreamType: streamType,
		Timestamp:  timestamp,
		Message:    message,
	}
}
