package s3

import (
	"log"
)

const DEFAULT_SECONDS_BETWEEN_POLLS = 5 * 60
const DEFAULT_PATHS_PER_BATCH = 100

type Options struct {
	CredsPath  string
	Region     string
	BucketName string
}

func ValidateOptions(options *Options) {
	if options.CredsPath == "" {
		log.Fatalf("Missing S3.CredsPath")
	}
	if options.Region == "" {
		log.Fatalf("Missing S3.Region")
	}
	if options.BucketName == "" {
		log.Fatalf("Missing S3.BucketName")
	}

}
