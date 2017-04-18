package main

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
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

type CreateRoleNonIdempotentReturn struct{}

func createRoleNonIdempotent(iamService *iam.IAM,
	roleName string) chan CreateRoleNonIdempotentReturn {

	future := make(chan CreateRoleNonIdempotentReturn)
	go func() {
		log.Printf("Creating role named %s...", roleName)

		output, err := iamService.CreateRole(&iam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Sid": "",
						"Effect": "Allow",
						"Principal": {
							"Service": "lambda.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					}
				]
			}`),
			Path:     aws.String("/"),
			RoleName: aws.String(roleName),
		})
		if err != nil {
			log.Fatalf("Error from CreateRole: %s", err)
		} else {
			log.Printf("Output from CreateRole: %s", output)
		}

		future <- CreateRoleNonIdempotentReturn{}
	}()
	return future
}

func getArnForRole(iamService *iam.IAM, roleName string) (string, error) {
	output, err := iamService.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return "", err
	} else {
		log.Printf("Output from GetRole: %s", output)
		return *output.Role.Arn, err
	}
}

type CreateRoleIdempotentReturn struct {
	arn string
}

func CreateRoleIdempotent(iamService *iam.IAM,
	roleName string) chan CreateRoleIdempotentReturn {
	future := make(chan CreateRoleIdempotentReturn)
	go func() {
		arn, err := getArnForRole(iamService, roleName)
		if err != nil {
			if err.Error() == iam.ErrCodeNoSuchEntityException {
				<-createRoleNonIdempotent(iamService, roleName)
				arn, err = getArnForRole(iamService, roleName)
				if err != nil {
					log.Fatalf("Error from second GetRole: %s", err)
				}
			} else {
				log.Fatalf("Error from first GetRole: %s", err)
			}
		}
		future <- CreateRoleIdempotentReturn{arn: arn}
	}()
	return future
}

func main() {
	sourceBucketName := "danstutzman-lambda-example"
	targetBucketName := sourceBucketName + "resized"
	functionName := "CreateThumbnail"
	executionRoleName := "lambda-" + functionName + "-execution"

	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	s3Service := s3.New(awsSession)
	iamService := iam.New(awsSession)

	if false {
		createSourceBucketFuture := CreateBucket(s3Service, sourceBucketName)
		createTargetBucketFuture := CreateBucket(s3Service, targetBucketName)

		<-createSourceBucketFuture
		copySampleImageFuture := CopyToBucket(s3Service, "../HappyFace.jpg",
			sourceBucketName, "/HappyFace.jpg")

		<-createTargetBucketFuture
		<-copySampleImageFuture
	}

	executionRoleArn := <-CreateRoleIdempotent(iamService, executionRoleName)
	_ = executionRoleArn
}