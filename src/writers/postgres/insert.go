package postgres

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

const MAX_COLUMNS_TO_ADD = 20

var MISSING_COLUMN_REGEXP = regexp.MustCompile(
	`pq: column "([^"]*)" of relation "([^"]*)" does not exist`)

func QuoteString(input string) string {
	input = strings.Replace(input, "'", "''", -1)
	return "'" + input + "'"
}

func (conn *Connection) InsertMaps(maps []map[string]interface{}) {
	keys := []string{"time"}
	for key := range maps[0] {
		if key == "timestamp" {
			// Skip it
		} else {
			keys = append(keys, key)
		}
	}

	tuples := []string{}
	for _, m := range maps {
		values := []string{}
		for _, key := range keys {
			value := m[key]
			if key == "time" {
				timeTime := m["timestamp"].(time.Time).UTC()
				values = append(values,
					"'"+timeTime.Format(time.RFC3339Nano)+"'")
			} else if valueFloat, ok := value.(float64); ok {
				values = append(values, fmt.Sprintf("%f", valueFloat))
			} else if valueString, ok := value.(string); ok {
				if key == "sql" && len(valueString) > 250 {
					valueString = valueString[0:250] + "..."
				}
				values = append(values, QuoteString(valueString))
			} else if valueInt, ok := value.(int64); ok {
				values = append(values, fmt.Sprintf("%d", valueInt))
			} else if valueBool, ok := value.(bool); ok {
				if valueBool {
					values = append(values, "'t'")
				} else {
					values = append(values, "'f'")
				}
			} else {
				log.Fatalw("Unexpected type for key", "key", key,
					"value", value, "type", fmt.Sprintf("%T", value))
			}
		}

		tuples = append(tuples, "("+strings.Join(values, ",")+")")
	}

	insertCommand := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		conn.tableName,
		strings.Join(keys, ", "),
		strings.Join(tuples, ","))

	for i := 1; i <= MAX_COLUMNS_TO_ADD; i++ {
		_, insertErr := conn.db.Exec(insertCommand)
		if insertErr != nil {
			match := MISSING_COLUMN_REGEXP.FindStringSubmatch(insertErr.Error())
			if match != nil {
				columnName := match[1]
				columnType := inferColumnType(maps, columnName)
				alterCommand := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
					conn.tableName, columnName, columnType)
				_, alterErr := conn.db.Exec(alterCommand)
				if alterErr != nil {
					log.Fatalw("Error from db.Exec", "sql", alterCommand, "err", insertErr)
					break
				}
			} else {
				log.Fatalw("Error from db.Exec", "sql", insertCommand, "err", insertErr)
				break // never executed
			}
		}
	}
}

func inferColumnType(maps []map[string]interface{}, columnName string) string {
	for _, m := range maps {
		value := m[columnName]
		if value != nil {
			if _, ok := value.(float64); ok {
				return "FLOAT"
			} else if _, ok := value.(string); ok {
				return "TEXT"
			} else if _, ok := value.(int64); ok {
				return "INT"
			} else if _, ok := value.(bool); ok {
				return "BOOL"
			} else {
				log.Fatalw("Unknown column type", "columnName", columnName)
			}
		}
	}
	log.Fatalw("Can't find row with column name", "columnName", columnName)
	return "" // Never executed
}
