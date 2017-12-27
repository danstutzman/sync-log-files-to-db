package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	_ "github.com/lib/pq"
)

type Connection struct {
	db        *sql.DB
	tableName string
}

func escape(s string) string {
	return strings.Replace(s, "'", "\\'", -1)
}

func NewConnection(opts *Options, configPath string) *Connection {
	parts := []string{
		fmt.Sprintf("dbname='%s'", escape(opts.DatabaseName)),
	}
	if opts.Hostname != nil {
		parts = append(parts, fmt.Sprintf("host='%s'", escape(*opts.Hostname)))
	}
	if opts.Port != nil {
		parts = append(parts, fmt.Sprintf("port='%s'", escape(*opts.Port)))
	}
	if opts.Username != nil {
		parts = append(parts, fmt.Sprintf("user='%s'", escape(*opts.Username)))
	}
	if opts.Password != nil {
		parts = append(parts, fmt.Sprintf("password='%s'", escape(*opts.Password)))
	}
	if opts.SslMode != nil {
		parts = append(parts, fmt.Sprintf("sslmode='%s'", escape(*opts.SslMode)))
	}

	db, err := sql.Open("postgres", strings.Join(parts, " "))
	if err != nil {
		log.Fatalw("Error from sql.Open", "err", err)
	}

	return &Connection{
		db:        db,
		tableName: opts.TableName,
	}
}
