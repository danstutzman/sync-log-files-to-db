package influxdb

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	clientPkg "github.com/influxdata/influxdb/client/v2"
)

func (conn *InfluxdbConnection) QueryForLastTimestamp(containerId string) time.Time {
	command := fmt.Sprintf(
		"SELECT LAST(message) FROM %s WHERE container_id = '%s'",
		conn.measurementName,
		strings.Replace(containerId, "\"", "\\\"", -1))

	results, err := conn.query(command)
	if err != nil {
		log.Fatalf("Error from query %s: %s", command, err)
	}

	if len(results) == 0 {
		return time.Unix(0, 0)
	} else if len(results) > 1 {
		log.Fatalf("Expected one or zero results but got %d", len(results))
	}

	if len(results[0].Series) == 0 {
		// No rows so return earliest possible time
		return time.Unix(0, 0)
	} else if len(results[0].Series) > 1 {
		log.Fatalf("Expected one or zero for len(Series) but got %d", len(results[0].Series))
	}
	row := results[0].Series[0]

	if len(row.Values) != 1 {
		log.Fatalf("Expected series.Values to be 1 but was %d", len(row.Values))
	}
	values := row.Values[0]

	for columnNum, columnName := range row.Columns {
		if columnName == "time" {
			nanos, err := values[columnNum].(json.Number).Int64()
			if err != nil {
				log.Fatalf("Can't convert %v to Int64", values[columnNum])
			}
			return time.Unix(0, nanos).UTC()
		}
	}
	log.Fatalf("Couldn't find time column in query result %v", results)
	return time.Unix(0, 0).UTC() // this line is never reached
}

func (conn *InfluxdbConnection) query(command string) (result []clientPkg.Result, err error) {

	q := clientPkg.Query{
		Command:   command,
		Database:  conn.databaseName,
		Precision: "ns",
	}
	if response, err := conn.client.Query(q); err == nil {
		if response.Error() != nil {
			return []clientPkg.Result{}, response.Error()
		}
		return response.Results, nil
	}
	return []clientPkg.Result{}, nil
}
