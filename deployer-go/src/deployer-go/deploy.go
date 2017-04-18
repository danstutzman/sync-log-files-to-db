package main

import (
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/deployer-go/src/lambda_deployer"
	"log"
	"os"
)

func main() {
	deployer := lambda_deployer.NewLambdaDeployer(lambda_deployer.Config{
		SourceBucketName: "danstutzman-lambda-example",
		TargetBucketName: "danstutzman-lambda-exampleresized",
		FunctionName:     "CreateThumbnail",
		RoleName:         "lambda-CreateThumbnail-execution",
	})
	if false {
		deployer.SetupBuckets()
	}
	deployer.DeployFunction()

	log.Printf("%s completed successfully", os.Args[0])
}
