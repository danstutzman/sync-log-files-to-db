package lambda_deployer

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"log"
)

func createBucket(s3Service *s3.S3, bucketName string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Creating bucket %s...", bucketName)
		output, err := s3Service.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			log.Fatalf("Error from CreateBucket: %s", err)
		}

		log.Printf("Output from CreateBucket: %s", output)

		future <- EmptyReturn{}
	}()
	return future
}

func copyToBucket(s3Service *s3.S3, fromPath string, toBucketName string,
	toPath string) chan EmptyReturn {
	future := make(chan EmptyReturn)
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

		future <- EmptyReturn{}
	}()
	return future
}
