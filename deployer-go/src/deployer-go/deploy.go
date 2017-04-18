package main

import (
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/deployer-go/src/aws_services"
	"log"
	"os"
)

func main() {
	sourceBucketName := "danstutzman-lambda-example"
	targetBucketName := sourceBucketName + "resized"
	functionName := "CreateThumbnail"
	roleName := "lambda-" + functionName + "-execution"

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

	roleArn := <-aws.CreateRoleIdempotent(roleName)
	_ = roleArn
	<-aws.PutRolePolicy(roleName, sourceBucketName, targetBucketName)

	log.Printf("%s completed successfully", os.Args[0])
}
