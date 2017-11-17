package s3_belugacdn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"time"

	bigquery "github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
	googleBigqueryPkg "google.golang.org/api/bigquery/v2"
)

var VISITS_TAG_SET = map[string]bool{
	"host": true,
}

var INTEGER_REGEXP = regexp.MustCompile("^[0-9]+$")
var FLOAT_REGEXP = regexp.MustCompile("^[0-9]+\\.[0-9]+$")

func readJsonIntoVisitMaps(reader io.Reader) []map[string]interface{} {
	visitMaps := []map[string]interface{}{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		visit := map[string]string{}
		err := json.Unmarshal([]byte(scanner.Text()), &visit)
		if err != nil {
			panic(fmt.Errorf("Error from json.Unmarshal: %s", err))
		}

		timestamp := time.Unix(toInt("time", visit["time"]), 0)

		visitMap := map[string]interface{}{
			"timestamp":          timestamp,
			"duration":           toFloat("duration", visit["duration"]),
			"response_size":      toInt("response_size", visit["response_size"]),
			"header_size":        toInt("header_size", visit["header_size"]),
			"trace":              visit["trace"],
			"server_region":      visit["server_region"],
			"protocol":           visit["protocol"],
			"property_name":      visit["property_name"],
			"status":             visit["status"], // don't insert as int
			"remote_addr":        visit["remote_addr"],
			"request_method":     visit["request_method"],
			"uri":                visit["uri"],
			"user_agent":         visit["user_agent"],
			"referer":            visit["referer"],
			"content_type":       visit["content_type"],
			"cache_status":       visit["cache_status"],
			"geo_continent":      visit["geo_continent"],
			"geo_continent_code": visit["geo_continent_code"],
			"geo_country":        visit["geo_country"],
			"geo_country_code":   visit["geo_country_code"],
		}

		visitMaps = append(visitMaps, visitMap)
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("Error from scanner.Err: %s", err))
	}

	return visitMaps
}

func toInt(key, value string) int64 {
	if !INTEGER_REGEXP.MatchString(value) {
		log.Fatalf("Unexpected characters in field %s: '%s'", key, value)
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatalf("Error from Atoi: %s", err)
	}
	return i
}

func toFloat(key, value string) float64 {
	if !FLOAT_REGEXP.MatchString(value) {
		log.Fatalf("Unexpected characters in field %s: '%s'", key, value)
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatalf("Error from ParseFloat: %s", err)
	}
	return f
}

func createVisitsTable(bigqueryConn *bigquery.BigqueryConnection) {
	bigqueryConn.CreateTable("visits", []*googleBigqueryPkg.TableFieldSchema{
		{Name: "time", Type: "TIMESTAMP", Mode: "REQUIRED"},
		{Name: "duration", Type: "FLOAT", Mode: "REQUIRED"},
		{Name: "trace", Type: "STRING", Mode: "REQUIRED"},
		{Name: "server_region", Type: "STRING", Mode: "REQUIRED"},
		{Name: "protocol", Type: "STRING", Mode: "REQUIRED"},
		{Name: "property_name", Type: "STRING", Mode: "NULLABLE"},
		{Name: "status", Type: "STRING", Mode: "REQUIRED"},
		{Name: "response_size", Type: "INTEGER", Mode: "REQUIRED"},
		{Name: "header_size", Type: "INTEGER", Mode: "REQUIRED"},
		{Name: "remote_addr", Type: "STRING", Mode: "REQUIRED"},
		{Name: "request_method", Type: "STRING", Mode: "REQUIRED"},
		{Name: "host", Type: "STRING", Mode: "REQUIRED"},
		{Name: "uri", Type: "STRING", Mode: "REQUIRED"},
		{Name: "user_agent", Type: "STRING", Mode: "REQUIRED"},
		{Name: "cache_status", Type: "STRING", Mode: "REQUIRED"},
		{Name: "geo_continent", Type: "STRING", Mode: "REQUIRED"},
		{Name: "geo_continent_code", Type: "STRING", Mode: "REQUIRED"},
		{Name: "geo_country", Type: "STRING", Mode: "REQUIRED"},
		{Name: "geo_country_code", Type: "STRING", Mode: "REQUIRED"},
	})
}
