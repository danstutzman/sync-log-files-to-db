package main

import (
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/deployer-go/src/lambda_deployer"
	"log"
	"os"
)

func main() {
	sourceBucketName := "danstutzman-lambda-example"
	targetBucketName := sourceBucketName + "resized"
	functionName := "CreateThumbnail"
	roleName := "lambda-" + functionName + "-execution"

	deployer := lambda_deployer.NewLambdaDeployer()

	if false {
		createSourceBucketFuture := deployer.CreateBucket(sourceBucketName)
		createTargetBucketFuture := deployer.CreateBucket(targetBucketName)

		<-createSourceBucketFuture
		copySampleImageFuture := deployer.CopyToBucket("../HappyFace.jpg",
			sourceBucketName, "/HappyFace.jpg")

		<-createTargetBucketFuture
		<-copySampleImageFuture
	}

	roleArn := <-deployer.CreateRoleIdempotent(roleName)
	_ = roleArn
	<-deployer.PutRolePolicy(roleName, sourceBucketName, targetBucketName)

	log.Printf("%s completed successfully", os.Args[0])
}
