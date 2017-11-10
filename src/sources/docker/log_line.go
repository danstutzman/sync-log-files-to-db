package docker

import (
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
	StreamType string
	Timestamp  time.Time
	Message    string
}

func tailLogLines(out io.Reader) chan *LogLine {
	c := make(chan *LogLine)
	go func() {
		for {
			logLine := readLogLineBlocking(out)
			if logLine == nil {
				close(c)
				break
			} else {
				c <- logLine
			}
		}
	}()
	return c
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
		panic(err)
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
		panic(err)
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
