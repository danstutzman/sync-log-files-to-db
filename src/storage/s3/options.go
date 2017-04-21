package s3

import (
	"log"
)

type Options struct {
	CredsPath  string
	Region     string
	BucketName string
}

func Usage() string {
	return `{
      "CredsPath":     STRING,  path to AWS credentials file, e.g. "./s3.creds.ini"
      "Region":        STRING,  AWS region for S3, e.g. "us-east-1"
      "BucketName":    STRING,  Name of S3 bucket, e.g. "cloudfront-logs-danstutzman"
    }`
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
