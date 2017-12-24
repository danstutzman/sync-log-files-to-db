package postgres

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

type Options struct {
	Hostname     *string
	Port         *string
	SslMode      *string
	Username     *string
	Password     *string
	DatabaseName string
	TableName    string
}

func ValidateOptions(options *Options) {
	if options.DatabaseName == "" {
		log.Fatalw("Missing Postgresql.DatabaseName")
	}
	if options.TableName == "" {
		log.Fatalw("Missing Postgresql.TableName")
	}
}
