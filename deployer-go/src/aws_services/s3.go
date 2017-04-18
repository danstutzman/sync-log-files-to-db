package aws_services

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"log"
)

type CreateBucketReturn struct{}

func (awsServices *AwsServices) CreateBucket(bucketName string) chan CreateBucketReturn {
	future := make(chan CreateBucketReturn)
	go func() {
		log.Printf("Creating bucket %s...", bucketName)
		output, err := awsServices.s3Service.CreateBucket(&s3.CreateBucketInput{
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

func (awsServices *AwsServices) CopyToBucket(fromPath string, toBucketName string,
	toPath string) chan CopyToBucketReturn {
	future := make(chan CopyToBucketReturn)
	go func() {
		log.Printf("Copying %s to bucket %s...", fromPath, toBucketName)

		fromBytes, err := ioutil.ReadFile(fromPath)
		if err != nil {
			log.Fatalf("Error from ReadFile: %s", err)
		}

		output, err := awsServices.s3Service.PutObject(&s3.PutObjectInput{
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
