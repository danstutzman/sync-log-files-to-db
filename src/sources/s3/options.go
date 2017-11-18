package s3

import (
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

type Options struct {
	CredsPath  string
	Region     string
	BucketName string
}

func ValidateOptions(options *Options) {
	if options.CredsPath == "" {
		log.Fatalw("Missing S3.CredsPath")
	}
	if options.Region == "" {
		log.Fatalw("Missing S3.Region")
	}
	if options.BucketName == "" {
		log.Fatalw("Missing S3.BucketName")
	}
}
