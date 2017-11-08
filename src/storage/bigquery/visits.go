package bigquery

import (
	bigquery "google.golang.org/api/bigquery/v2"
)

func maybeNull(s string) bigquery.JsonValue {
	if s == "-" {
		return nil
	} else if s == "" {
		return nil
	} else {
		return s
	}
}

func (bigqueryConn *BigqueryConnection) CreateVisitsTable() {
	bigqueryConn.CreateTable("visits", []*bigquery.TableFieldSchema{
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

func (bigqueryConn *BigqueryConnection) UploadVisits(visits []map[string]string) {

	rows := make([]*bigquery.TableDataInsertAllRequestRows, 0)
	for _, visit := range visits {
		row := &bigquery.TableDataInsertAllRequestRows{
			InsertId: visit["trace"],
			Json: map[string]bigquery.JsonValue{
				"time":               visit["time"],
				"duration":           visit["duration"],
				"trace":              visit["trace"],
				"server_region":      visit["server_region"],
				"protocol":           visit["protocol"],
				"property_name":      visit["property_name"],
				"status":             visit["status"],
				"response_size":      visit["response_size"],
				"header_size":        visit["header_size"],
				"remote_addr":        visit["remote_addr"],
				"request_method":     visit["request_method"],
				"host":               visit["host"],
				"uri":                visit["uri"],
				"user_agent":         visit["user_agent"],
				"cache_status":       visit["cache_status"],
				"geo_continent":      visit["geo_continent"],
				"geo_continent_code": visit["geo_continent_code"],
				"geo_country":        visit["geo_country"],
				"geo_country_code":   visit["geo_country_code"],
			}}
		rows = append(rows, row)
	}

	bigqueryConn.InsertRows("visits",
		func() { bigqueryConn.CreateDataset() },
		func() { bigqueryConn.CreateVisitsTable() },
		rows)
}
