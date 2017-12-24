package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

func quoteString(input string) string {
	return fmt.Sprintf("'%s'", strings.Replace(input, "'", "\\'", -1))
}

func (conn *PostgresConnection) InsertMaps(maps []map[string]interface{}) {
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
				values = append(values, quoteString(valueString))
			} else if valueInt, ok := value.(int64); ok {
				values = append(values, fmt.Sprintf("%d", valueInt))
			} else {
				log.Fatalw("Unexpected type for key", "key", key,
					"value", value, "type", fmt.Sprintf("%T", value))
			}
		}

		tuples = append(tuples, "("+strings.Join(values, ",")+")")
	}
	command := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON CONFLICT DO NOTHING",
		conn.tableName,
		strings.Join(keys, ", "),
		strings.Join(tuples, ","))

	_, err := conn.db.Exec(command)
	if err != nil {
		log.Fatalw("Error from db.Exec", "sql", command, "err", err)
	}
}
