package postgres

import (
	"fmt"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

func (conn *PostgresConnection) CreateBelugacdnLogsTable() {
	command := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		time               TIMESTAMP NOT NULL UNIQUE,
		header_size        INT NOT NULL,
		status             TEXT NOT NULL,
		cache_status       TEXT NOT NULL,
		geo_continent_code TEXT NOT NULL,
		property_name      TEXT NOT NULL,
		geo_continent      TEXT NOT NULL,
		trace              TEXT NOT NULL,
		response_size      INT NOT NULL,
		user_agent         TEXT NOT NULL,
		uri                TEXT NOT NULL,
		duration           FLOAT NOT NULL,
		remote_addr        TEXT NOT NULL,
		referer            TEXT NOT NULL,
		request_method     TEXT NOT NULL,
		content_type       TEXT NOT NULL,
		server_region      TEXT NOT NULL,
		host               TEXT NOT NULL,
		geo_country        TEXT NOT NULL,
		geo_country_code   TEXT NOT NULL,
		protocol           TEXT NOT NULL
	)`, conn.tableName)

	_, err := conn.db.Exec(command)
	if err != nil {
		log.Fatalw("Error from db.Exec", "sql", command, "err", err)
	}
}

func (conn *PostgresConnection) CreateMonitisResultsTable() {
	command1 := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		time               TIMESTAMP NOT NULL,
		monitor_name       TEXT NOT NULL,
		location_name      TEXT NOT NULL,
		response_millis    INT NOT NULL,
		was_okay           BOOL NOT NULL
	)`, conn.tableName)

	_, err := conn.db.Exec(command1)
	if err != nil {
		log.Fatalw("Error from db.Exec", "sql", command1, "err", err)
	}

	command2 := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS
		idx_%s_time_monitor_name_location_name
		ON %s(monitor_name, location_name, time)`,
		conn.tableName, conn.tableName)
	_, err = conn.db.Exec(command2)
	if err != nil {
		log.Fatalw("Error from db.Exec", "sql", command2, "err", err)
	}
}
