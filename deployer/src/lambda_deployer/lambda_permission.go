package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"log"
)

func addPermission(lambdaService *lambda.Lambda, functionName string, sourceBucketName string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Adding permission for %s...", functionName)
		output, err := lambdaService.AddPermission(&lambda.AddPermissionInput{
			Action:       aws.String("lambda:InvokeFunction"),
			FunctionName: aws.String(functionName),
			Principal:    aws.String("s3.amazonaws.com"),
			StatementId:  aws.String("some-unique-id"),
			SourceArn:    aws.String("arn:aws:s3:::" + sourceBucketName),
		})

		if err != nil {
			if err2, ok := err.(awserr.Error); ok && err2.Code() ==
				lambda.ErrCodeResourceConflictException {
				// already exists, so ignore
				future <- EmptyReturn{}
			} else {
				log.Fatalf("Error from AddPermission: %s", err)
			}
		} else {
			log.Printf("Output from AddPermission: %v", output)
			future <- EmptyReturn{}
		}
	}()
	return future
}
