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
