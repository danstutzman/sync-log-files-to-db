package main

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"log"
)

type CreateBucketReturn struct{}

func CreateBucket(s3Service *s3.S3, bucketName string) chan CreateBucketReturn {
	future := make(chan CreateBucketReturn)
	go func() {
		log.Printf("Creating bucket %s...", bucketName)
		output, err := s3Service.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			log.Fatalf("Error from CreateBucket: %s", err)
		}

		log.Printf("Output from CreateBucket: %s", output)

		future <- CreateBucketReturn{}
	}()
	return future
}

type CopyToBucketReturn struct{}

func CopyToBucket(s3Service *s3.S3, fromPath string, toBucketName string,
	toPath string) chan CopyToBucketReturn {
	future := make(chan CopyToBucketReturn)
	go func() {
		log.Printf("Copying %s to bucket %s...", fromPath, toBucketName)

		fromBytes, err := ioutil.ReadFile(fromPath)
		if err != nil {
			log.Fatalf("Error from ReadFile: %s", err)
		}

		output, err := s3Service.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(toBucketName),
			Key:    aws.String(toPath),
			Body:   bytes.NewReader(fromBytes),
		})
		if err != nil {
			log.Fatalf("Error from PutObject: %s", err)
		}

		log.Printf("Output from PutObject: %s", output)

		future <- CopyToBucketReturn{}
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
		createSourceBucketFuture := CreateBucket(s3Service, sourceBucketName)
		createTargetBucketFuture := CreateBucket(s3Service, targetBucketName)

		copySampleImageFuture := CopyToBucket(s3Service, "../HappyFace.jpg",
			sourceBucketName, "/HappyFace.jpg")

	<-createSourceBucketFuture

	<-createTargetBucketFuture
	<-copySampleImageFuture
}
