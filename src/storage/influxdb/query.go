package influxdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	clientPkg "github.com/influxdata/influxdb/client/v2"
)

func (conn *InfluxdbConnection) QueryForLastTimestampForTag(tagName, tagValue string) time.Time {
	command := fmt.Sprintf(
		"SELECT LAST(message) FROM %s WHERE %s = '%s'",
		conn.measurementName,
		tagName,
		strings.Replace(tagValue, "\"", "\\\"", -1))
	return conn.queryForLastTimestamp(command)
}

func (conn *InfluxdbConnection) QueryForLastTimestamp() time.Time {
	command := fmt.Sprintf("SELECT LAST(message) FROM %s", conn.measurementName)
	return conn.queryForLastTimestamp(command)
}

func (conn *InfluxdbConnection) queryForLastTimestamp(command string) time.Time {

	results, err := conn.query(command)
	if err != nil {
		log.Fatalw("Error from query", "query", command, "err", err)
	}

	if len(results) == 0 {
		return time.Time{} // zero value
	} else if len(results) > 1 {
		log.Fatalw("Expected one or zero results", "got", len(results))
	}

	if len(results[0].Series) == 0 {
		// No rows so return earliest possible time
		return time.Time{} // zero value
	} else if len(results[0].Series) > 1 {
		log.Fatalw("Expected one or zero for len(Series)", "got", len(results[0].Series))
	}
	row := results[0].Series[0]

	if len(row.Values) != 1 {
		log.Fatalw("Expected series.Values to be 1", "got", len(row.Values))
	}
	values := row.Values[0]

	for columnNum, columnName := range row.Columns {
		if columnName == "time" {
			nanos, err := values[columnNum].(json.Number).Int64()
			if err != nil {
				log.Fatalw("Can't convert to Int64", "got", values[columnNum])
			}
			return time.Unix(0, nanos).UTC()
		}
	}
	log.Fatalw("Couldn't find time column in query result", "results", results)
	return time.Time{} // zero value; this line is never reached
}

func (conn *InfluxdbConnection) query(command string) (result []clientPkg.Result, err error) {
	q := clientPkg.Query{
		Command:   command,
		Database:  conn.databaseName,
		Precision: "ns",
	}

	response, err := conn.client.Query(q)
	if err != nil {
		return []clientPkg.Result{}, err
	}

	if response.Error() != nil {
		return []clientPkg.Result{}, response.Error()
	}

	return response.Results, nil
}
