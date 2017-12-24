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
