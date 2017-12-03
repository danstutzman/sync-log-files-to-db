package systemd

import (
	"bufio"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
)

const MAX_INFLUXDB_INSERT_BATCH_SIZE = 100
const JOURNALCTL_TIME_FORMAT = "2006-01-02 15:04:05.999999999"
const JOURNALCTL_LOGS_PRELUDE = "-- Logs begin at "
const JOURNALCTL_LOGS_REBOOT = "-- Reboot --"

var TAIL_LOG_LINE_FLUSH_TIMEOUT = time.Second
var INFLUXDB_TAGS_SET = map[string]bool{
	"machine":   true,
	"unit_name": true,
}
var SHORT_UNIX_LINE_REGEXP = regexp.MustCompile(`^([0-9.]+) ([^ ]+) ([^ []*)(\[[^]]+\])?: (.*)$`)

type LogLine struct {
	Timestamp   time.Time
	Machine     string
	ProcessName string
	Message     string
}

func StartTailingSystemdLogs(config *Options, configPath string) {
	var influxdbConn *influxdb.InfluxdbConnection
	influxdb.ValidateOptions(config.InfluxDb)
	influxdbConn = influxdb.NewInfluxdbConnection(config.InfluxDb, configPath)
	influxdbConn.CreateDatabase()

	logLinesChan := make(chan LogLine)
	go startTailingSystemdLog(influxdbConn, logLinesChan)
	syncToInfluxdbForever(logLinesChan, influxdbConn, config.UnitNames)
}

func startTailingSystemdLog(influxdbConn *influxdb.InfluxdbConnection,
	logLinesChan chan<- LogLine) {

	lastTimestamp := influxdbConn.QueryForLastTimestamp()

	args := []string{
		"/usr/bin/journalctl",
		"--since=" + lastTimestamp.Format(JOURNALCTL_TIME_FORMAT),
		"--follow",
		"--no-pager",
		"--no-tail",
		"--output=short-unix",
	}
	log.Infow("Tailing logs...", "after", lastTimestamp, "args", args)

	command := exec.Command(args[0], args[1:len(args)]...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		log.Fatalw("Error from StdoutPipe", "err", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		log.Fatalw("Error from StderrPipe", "err", err)
	}
	err = command.Start()
	if err != nil {
		log.Fatalw("Error from Start", "err", err)
	}

	go tailSystemdLog(stdout, influxdbConn, logLinesChan)

	stderrOut, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Fatalw("Error from ReadAll", "err", err)
	}

	err2 := command.Wait()
	if err2 != nil {
		log.Fatalw("Error from Wait", "err", err, "stderr", string(stderrOut))
	}

}

func tailSystemdLog(stdout io.Reader,
	influxdbConn *influxdb.InfluxdbConnection,
	logLinesChan chan<- LogLine) {

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, JOURNALCTL_LOGS_PRELUDE) {
			// ignore it
		} else if strings.HasPrefix(line, JOURNALCTL_LOGS_REBOOT) {
			// ignore it
		} else {
			match := SHORT_UNIX_LINE_REGEXP.FindStringSubmatch(scanner.Text())
			if match == nil {
				log.Fatalw("Line doesn't match SHORT_UNIX_LINE_REGEXP", "line", line)
			}

			secondsSinceEpoch, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				log.Fatalw("Error from ParseFloat", "input", match[1])
			}

			logLinesChan <- LogLine{
				Timestamp:   time.Unix(0, int64(secondsSinceEpoch*1000000000)),
				Machine:     match[2],
				ProcessName: match[3],
				Message:     match[5],
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalw("Error from scanner.Err", "err", err)
	}
}

func syncToInfluxdbForever(logLinesChan <-chan LogLine,
	influxdbConn *influxdb.InfluxdbConnection, unitNames []string) {

	maps := []map[string]interface{}{}
	for {
		if len(maps) == 0 {
			logLine := <-logLinesChan
			maps = appendLogLineToMaps(logLine, maps, unitNames)
		} else if len(maps) >= MAX_INFLUXDB_INSERT_BATCH_SIZE {
			influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, maps)
			maps = []map[string]interface{}{}
		} else {
			select {
			case logLine := <-logLinesChan:
				maps = appendLogLineToMaps(logLine, maps, unitNames)
			case <-time.After(TAIL_LOG_LINE_FLUSH_TIMEOUT):
				// No new logs after timeout, so inserting
				influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, maps)
				maps = []map[string]interface{}{}
			}
		}
	}
}

func appendLogLineToMaps(logLine LogLine,
	maps []map[string]interface{}, unitNames []string) []map[string]interface{} {

	unitName := "other"
	for _, givenUnitName := range unitNames {
		if logLine.ProcessName == givenUnitName {
			unitName = logLine.ProcessName
		}
	}

	logLineAsMap := map[string]interface{}{
		"timestamp":    logLine.Timestamp,
		"machine":      logLine.Machine,
		"process_name": logLine.ProcessName,
		"message":      logLine.Message,
		"unit_name":    unitName,
	}

	return append(maps, logLineAsMap)
}
