package main

import (
	"github.com/danielstutzman/sync-logs-from-s3/src/lambda_deployer"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 1+1 {
		log.Fatalf("Usage: " + os.Args[0] +
			" [remove-function] [install-function] [sync-existing-logs]")
	}
	verbs := os.Args[1:]

	deployer := lambda_deployer.NewLambdaDeployer(lambda_deployer.Config{
		BucketName:   "cloudfront-logs-danstutzman",
		FunctionName: "SyncCloudfrontLogsToBigquery",
		RoleName:     "lambda-SyncCloudfrontLogsToBigquery-execution",
		PolicyName:   "lambda-SyncCloudfrontLogsToBigquery-execution-access",
	})

	for _, verb := range verbs {
		switch verb {
		case "remove-function":
			deployer.RemoveFunction()
		case "install-function":
			deployer.InstallFunction()
		case "sync-existing-logs":
			deployer.SyncExistingLogs()
		default:
			log.Fatalf("Unknown verb '%s'; expected up or down", verb)
		}
	}

	log.Printf("%s completed successfully", os.Args[0])
}
