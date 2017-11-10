package s3

import (
	"fmt"
	"io"
	"log"
	"path"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cenkalti/backoff"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/bigquery"
)

const NUM_PATHS_PER_PAGE = 1000

type S3Connection struct {
	service    *s3.S3
	bucketName string
}

func PollForever(opts *Options, configPath string) {
	s3Conn := NewS3Connection(opts, configPath)

	var bigqueryConn *bigquery.BigqueryConnection
	if opts.BigQuery != nil {
		bigqueryConn = bigquery.NewBigqueryConnection(opts.BigQuery, configPath)
	}

	var influxdbConn *influxdb.InfluxdbConnection
	if opts.InfluxDb != nil {
		influxdbConn = influxdb.NewInfluxdbConnection(opts.InfluxDb, configPath)
		influxdbConn.CreateDatabase()
	}

	for {
		visits := []map[string]string{}
		s3Paths := s3Conn.ListPaths(NUM_PATHS_PER_PAGE)
		for _, s3Path := range s3Paths {
			visits = append(visits, s3Conn.DownloadVisitsForPath(s3Path)...)
		}

		if len(visits) > 0 {
			if bigqueryConn != nil {
				bigqueryConn.UploadVisits(visits)
			}
			if influxdbConn != nil {
				influxdbConn.InsertVisits(visits)
			}
		}
		for _, s3Path := range s3Paths {
			s3Conn.DeletePath(s3Path)
		}

		log.Printf("Sleeping 5 minutes while waiting for next batch of logs")
		time.Sleep(5 * time.Minute)
	}
}

func NewS3Connection(opts *Options, configPath string) *S3Connection {
	log.Printf("Creating AWS session...")
	config := aws.Config{
		Region: aws.String(opts.Region),
	}
	if opts.CredsPath != "" {
		credsPath := path.Join(path.Dir(configPath), opts.CredsPath)
		config.Credentials = credentials.NewSharedCredentials(credsPath, "")
	}
	session, err := session.NewSessionWithOptions(session.Options{
		Config: config,
	})
	if err != nil {
		panic(fmt.Errorf("Couldn't create AWS session: %s", err))
	}

	return &S3Connection{
		service:    s3.New(session),
		bucketName: opts.BucketName,
	}
}

func (conn *S3Connection) ListPaths(maxKeys int64) []string {
	var response *s3.ListObjectsV2Output
	var err error

	err = backoff.Retry(func() error {
		log.Printf("Listing objects in s3://%s...", conn.bucketName)
		response, err = conn.service.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:  aws.String(conn.bucketName),
			MaxKeys: aws.Int64(maxKeys),
		})
		if err != nil {
			err2, isRequestFailure := err.(awserr.RequestFailure)
			if !isRequestFailure {
				log.Fatalf("Error from ListObjectsV2(Bucket=%s): %s", conn.bucketName, err)
			} else if err2.StatusCode() == 500 || err2.StatusCode() == 503 {
				// Let the backoff library retry
			} else {
				log.Fatalf("Error from ListObjectsV2(Bucket=%s): %s", conn.bucketName, err2)
			}
		}
		return err
	}, backoff.NewExponentialBackOff())
	if err != nil {
		err2, isRequestFailure := err.(awserr.RequestFailure)
		if !isRequestFailure {
			log.Fatalf("Error from ListObjectsV2(Bucket=%s): %s", conn.bucketName, err)
		} else {
			log.Fatalf("Error from ListObjectsV2(Bucket=%s): %s", conn.bucketName, err2)
		}
	}

	paths := []string{}
	for _, object := range response.Contents {
		paths = append(paths, *object.Key)
	}
	return paths
}

func (conn *S3Connection) DownloadPath(path string) io.ReadCloser {
	var response *s3.GetObjectOutput
	var err error

	err = backoff.Retry(func() error {
		log.Printf("Downloading %s...", path)
		response, err = conn.service.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(conn.bucketName),
			Key:    aws.String(path),
		})
		if err != nil {
			err2, isRequestFailure := err.(awserr.RequestFailure)
			if !isRequestFailure {
				log.Fatalf("Error from GetObject(Bucket=%s,Key=%s): %s",
					conn.bucketName, path, err)
			} else if err2.StatusCode() == 500 || err2.StatusCode() == 503 {
				// Let the backoff library retry
			} else {
				log.Fatalf("Error from GetObject(Bucket=%s,Key=%s): %s",
					conn.bucketName, path, err2)
			}
		}
		return err
	}, backoff.NewExponentialBackOff())

	if err != nil {
		err2, isRequestFailure := err.(awserr.RequestFailure)
		if !isRequestFailure {
			log.Fatalf("Error from GetObject(Bucket=%s,Key=%s): %s",
				conn.bucketName, path, err)
		} else {
			log.Fatalf("Error from GetObject(Bucket=%s,Key=%s): %s",
				conn.bucketName, path, err2)
		}
	}

	return response.Body
}

func (conn *S3Connection) DeletePath(path string) {
	err := backoff.Retry(func() error {
		log.Printf("Deleting %s...", path)
		_, err := conn.service.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(conn.bucketName),
			Key:    aws.String(path),
		})
		if err != nil {
			err2, isRequestFailure := err.(awserr.RequestFailure)
			if !isRequestFailure {
				log.Fatalf("Error from DeleteObject(Bucket=%s,Key=%s): %s",
					conn.bucketName, path, err)
			} else if err2.StatusCode() == 500 || err2.StatusCode() == 503 {
				// Let the backoff library retry
			} else {
				log.Fatalf("Error from DeleteObject(Bucket=%s,Key=%s): %s",
					conn.bucketName, path, err2)
			}
		}
		return err
	}, backoff.NewExponentialBackOff())

	if err != nil {
		err2, isRequestFailure := err.(awserr.RequestFailure)
		if !isRequestFailure {
			log.Fatalf("Error from DeleteObject(Bucket=%s,Key=%s): %s",
				conn.bucketName, path, err)
		} else {
			log.Fatalf("Error from DeleteObject(Bucket=%s,Key=%s): %s",
				conn.bucketName, path, err2)
		}
	}
}
