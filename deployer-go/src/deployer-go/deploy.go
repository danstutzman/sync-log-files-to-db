package main

import (
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/deployer-go/src/lambda_deployer"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 1+1 {
		log.Fatalf("Usage: " + os.Args[0] + " up|down")
	}
	verb := os.Args[1]

	deployer := lambda_deployer.NewLambdaDeployer(lambda_deployer.Config{
		SourceBucketName: "danstutzman-lambda-example",
		TargetBucketName: "danstutzman-lambda-exampleresized",
		FunctionName:     "CreateThumbnail",
		RoleName:         "lambda-CreateThumbnail-execution",
		PolicyName:       "lambda-CreateThumbnail-execution-access",
	})

	switch verb {
	case "down":
		deployer.DeleteEverything()
	case "up":
		deployer.SetupBuckets()
		deployer.DeployFunction()
	default:
		log.Fatalf("Unknown verb '%s'; expected up or down", verb)
	}

	log.Printf("%s completed successfully", os.Args[0])
}
