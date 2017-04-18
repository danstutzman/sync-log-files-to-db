package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"log"
)

func deleteFunction(lambdaService *lambda.Lambda, functionName string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Deleting function named %s...", functionName)
		_, err := lambdaService.DeleteFunction(&lambda.DeleteFunctionInput{
			FunctionName: aws.String(functionName),
		})

		if err != nil {
			if err2, ok := err.(awserr.Error); ok && err2.Code() == lambda.ErrCodeResourceNotFoundException {
				// ignore error
				future <- EmptyReturn{}
			} else {
				log.Fatalf("Error from DeleteFunction: %s", err)
			}
		} else {
			future <- EmptyReturn{}
		}
	}()
	return future
}
