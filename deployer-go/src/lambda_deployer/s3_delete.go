package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func deletePageOfObjects(s3Service *s3.S3, bucketName string) {
	log.Printf("Listing bucket named %s...", bucketName)
	output, err := s3Service.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Fatalf("Error from ListBucket: %s", err)
	} else {
		objectsToDelete := []*s3.ObjectIdentifier{}
		for _, object := range output.Contents {
			objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
				Key: object.Key,
			})
		}

		log.Printf("Deleting page of objects...")
		_, err := s3Service.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3.Delete{
				Objects: objectsToDelete,
			},
		})
		if err != nil {
			log.Fatalf("Error from DeleteObjects: %s", err)
		} else {
			return
		}
	}
}

func deleteBucket(s3Service *s3.S3, bucketName string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Deleting bucket named %s...", bucketName)
		_, err := s3Service.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			if err2, ok := err.(awserr.Error); ok && err2.Code() ==
				s3.ErrCodeNoSuchBucket {
				// ignore error
				future <- EmptyReturn{}
			} else if err2, ok := err.(awserr.Error); ok && err2.Code() ==
				"BucketNotEmpty" {
				deletePageOfObjects(s3Service, bucketName)
				<-deleteBucket(s3Service, bucketName) // try again
			} else {
				log.Fatalf("Error from DeleteBucket: %s", err)
			}
		} else {
			future <- EmptyReturn{}
		}
	}()
	return future
}
