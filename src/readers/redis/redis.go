package redis

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

var gotAuth = false
var DOLLAR_INT_REGEXP = regexp.MustCompile("\\$([0-9]+)\r")
var REDIS_KEY_NAME = "belugacdn"
var ASCII_CR = byte(13)
var ASCII_LF = byte(10)
var INFLUXDB_TAGS_SET = map[string]bool{"image_name": true}

func awaitAuthCommand(reader *bufio.Reader, conn net.Conn, expectedPassword string) {
	log.Infow("Awaiting AUTH command...")
	expect(reader, "*2")                                      // AUTH command has 2 parts
	expect(reader, "$4")                                      // part 1 has 4 chars
	expect(reader, "AUTH")                                    // part 1 is the word AUTH
	expect(reader, fmt.Sprintf("$%d", len(expectedPassword))) // part 2 has n chars
	expect(reader, expectedPassword)                          // part 2 is the password

	_, err := conn.Write([]byte("+OK\r\n"))
	if err != nil {
		log.Fatalw("Error from Write", "err", err)
	}
}

func awaitLpushCommand(reader *bufio.Reader, conn net.Conn,
	influxdbConn *influxdb.InfluxdbConnection, config *Options) {

	log.Infow("Awaiting LPUSH command...")
	expect(reader, "*3")                                    // LPUSH command has 3 parts
	expect(reader, "$5")                                    // part 1 has 5 chars
	expect(reader, "LPUSH")                                 // part 1 is the word LPUSH
	expect(reader, fmt.Sprintf("$%d", len(REDIS_KEY_NAME))) // part 2 has n chars
	expect(reader, REDIS_KEY_NAME)                          // part 2 is the key

	var upcomingStringLength = expectDollarInt(reader)
	log.Infow("Got upcomingStringLength", "got", upcomingStringLength)

	var logJson = make([]byte, upcomingStringLength)
	var numBytesRead, err = reader.Read(logJson)
	if err != nil {
		log.Fatalw("Error from Read", "err", err)
	}
	if numBytesRead != upcomingStringLength {
		log.Fatalw("Unexpected number of bytes",
			"expected", upcomingStringLength, "got", numBytesRead)
	}
	log.Infow("Read log", "logJson", string(logJson))

	cr, err := reader.ReadByte()
	if cr != ASCII_CR {
		log.Fatalw("Expected CR", "got", cr)
	}
	lf, err := reader.ReadByte()
	if lf != ASCII_LF {
		log.Fatalw("Expected LF", "got", lf)
	}

	map1 := parseLogJson(logJson)

	map2 := map[string]interface{}{}
	for key, value := range map1 {
		valueString := value.(string)

		if key == "time" {
			numSecondsInt, err := strconv.Atoi(valueString)
			if err != nil {
				log.Fatalw("Error from Atoi", "key", key, "value", valueString)
			}
			map2["timestamp"] = time.Unix(int64(numSecondsInt), 0)
		} else if key == "response_size" || key == "header_size" {
			// Don't consider key=status to be an integer
			valueInt, err := strconv.Atoi(value.(string))
			if err != nil {
				log.Fatalw("Expected integer", "key", key, "value", valueString)
			}
			map2[key] = valueInt
		} else if key == "duration" {
			valueFloat, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				log.Fatalw("Expected float", "key", key, "value", valueString)
			}
			map2[key] = valueFloat
		} else {
			map2[key] = value
		}
	}

	influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, []map[string]interface{}{map2})

	_, err = conn.Write([]byte(":1\r\n")) // say the length of the list is 1 long
	if err != nil {
		log.Fatalw("Error from Write", "err", err)
	}
}

func parseLogJson(logJson []byte) map[string]interface{} {
	parsed := &map[string]interface{}{}
	err := json.Unmarshal(logJson, parsed)
	if err != nil {
		log.Fatalw("Error from Unmarshal", "err", err)
	}
	return *parsed
}

func handleConnection(conn net.Conn, config *Options, influxdbConn *influxdb.InfluxdbConnection) {
	log.Infow("Handling new connection...")

	// Close connection when this function ends
	defer func() {
		log.Infow("Closing connection...")
		conn.Close()
	}()

	reader := bufio.NewReader(conn)

	// Set a deadline for reading. Read operation will fail if no data
	// is received after deadline.
	// timeoutDuration := 5 * time.Second
	// conn.SetReadDeadline(time.Now().Add(timeoutDuration))

	awaitAuthCommand(reader, conn, config.ExpectedPassword)

	for {
		awaitLpushCommand(reader, conn, influxdbConn, config)
	}
}

func expect(reader *bufio.Reader, expected string) {
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		log.Fatalw("Error from ReadBytes", "err", err)
	}

	if strings.ToUpper(strings.TrimSpace(string(bytes))) != strings.ToUpper(expected) {
		log.Fatalw("Unexpected number of bytes", "expected", expected, "got", bytes)
	}
}

func expectDollarInt(reader *bufio.Reader) int {
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		log.Fatalw("Error from ReadBytes", "err", err)
	}
	var match = DOLLAR_INT_REGEXP.FindStringSubmatch(string(bytes))
	log.Infow("Got match", "match", match[1])

	i, err := strconv.Atoi(match[1])
	if err != nil {
		log.Fatalw("Error from Atoi", "err", err, "input", match[1])
	}

	return i
}

func startRedisListener(config *Options, influxdbConn *influxdb.InfluxdbConnection) {
	listener, err := net.Listen("tcp", ":"+config.ListenPort)
	if err != nil {
		log.Fatalw("Error from Listen", "err", err)
	}
	log.Infow("Listening...", "port", config.ListenPort)

	defer func() {
		listener.Close()
		log.Infow("Listener closed")
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalw("Error from Accept", "err", err)
		}
		go handleConnection(conn, config, influxdbConn)
	}
}

func ListenForever(config *Options, configPath string) {
	var influxdbConn *influxdb.InfluxdbConnection
	if config.InfluxDb != nil {
		influxdb.ValidateOptions(config.InfluxDb)
		influxdbConn = influxdb.NewInfluxdbConnection(config.InfluxDb, configPath)
	}
	influxdbConn.CreateDatabase()

	startRedisListener(config, influxdbConn)
}
