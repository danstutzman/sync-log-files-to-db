package main

import (
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/deployer-go/src/aws_services"
)

func main() {
	sourceBucketName := "danstutzman-lambda-example"
	targetBucketName := sourceBucketName + "resized"
	functionName := "CreateThumbnail"
	executionRoleName := "lambda-" + functionName + "-execution"

	aws := aws_services.NewAwsServices()

	if false {
		createSourceBucketFuture := aws.CreateBucket(sourceBucketName)
		createTargetBucketFuture := aws.CreateBucket(targetBucketName)

		<-createSourceBucketFuture
		copySampleImageFuture := aws.CopyToBucket("../HappyFace.jpg",
			sourceBucketName, "/HappyFace.jpg")

		<-createTargetBucketFuture
		<-copySampleImageFuture
	}

	executionRoleArn := <-aws.CreateRoleIdempotent(executionRoleName)
	_ = executionRoleArn
}
