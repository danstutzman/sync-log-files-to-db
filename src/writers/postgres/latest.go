package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

func (conn *PostgresConnection) QueryForLastTimestamp(whereClause string) time.Time {
	var maxTime *time.Time
	query := fmt.Sprintf(
		"SELECT MAX(time) AS max_time FROM %s WHERE %s", conn.tableName, whereClause)
	err := conn.db.QueryRow(query).Scan(&maxTime)

	if err == sql.ErrNoRows { // can't happen
		return time.Time{} // zero time
	} else if maxTime == nil {
		return time.Time{} // zero time
	} else if err != nil {
		log.Fatalw("Error from QueryRow", "query", query, "err", err)
		return time.Time{} // never reached
	} else {
		return *maxTime
	}
}
