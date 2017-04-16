package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

type CreateBucketReturn struct{}

func createBucket(s3Service *s3.S3, bucketName string) chan CreateBucketReturn {
	future := make(chan CreateBucketReturn)
	go func() {
		log.Printf("Creating bucket %s...", bucketName)
		output, err := s3Service.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			log.Fatalf("Error from CreateBucket: %s", err)
		}

		log.Println(output)

		future <- CreateBucketReturn{}
	}()
	return future
}

func main() {
	sourceBucketName := "danstutzman-lambda-example"
	targetBucketName := sourceBucketName + "resized"

	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	s3Service := s3.New(awsSession)

	createBucketFuture1 := createBucket(s3Service, sourceBucketName)
	createBucketFuture2 := createBucket(s3Service, targetBucketName)
	<-createBucketFuture1
	<-createBucketFuture2
}
