package postgres

import (
	"fmt"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

func (conn *PostgresConnection) CreateTable() {
	command1 := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		time TIMESTAMP WITH TIME ZONE NOT NULL
	)`, conn.tableName)

	_, err := conn.db.Exec(command1)
	if err != nil {
		log.Fatalw("Error from db.Exec", "sql", command1, "err", err)
	}

	command2 := fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS idx_%s_time ON %s(time)`,
		conn.tableName, conn.tableName)
	_, err = conn.db.Exec(command2)
	if err != nil {
		log.Fatalw("Error from db.Exec", "sql", command2, "err", err)
	}
}
