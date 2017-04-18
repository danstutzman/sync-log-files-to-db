package lambda_deployer

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"log"
)

type InvokeFunctionReturn struct {
	logText string
}

func invokeFunction(lambdaService *lambda.Lambda, functionName string, sourceBucketName string) chan InvokeFunctionReturn {
	future := make(chan InvokeFunctionReturn)
	go func() {
		log.Printf("Invoking function %s...", functionName)
		output, err := lambdaService.Invoke(&lambda.InvokeInput{
			FunctionName:   aws.String(functionName),
			InvocationType: aws.String("RequestResponse"),
			LogType:        aws.String("Tail"),
			Payload: []byte(`{
				 "Records":[
					 {
							"eventVersion":"2.0",
							"eventSource":"aws:s3",
							"awsRegion":"us-east-1",
							"eventTime":"1970-01-01T00:00:00.000Z",
							"eventName":"ObjectCreated:Put",
							"userIdentity":{
								 "principalId":"AIDAJDPLRKLG7UEXAMPLE"
							},
							"requestParameters":{
								 "sourceIPAddress":"127.0.0.1"
							},
							"responseElements":{
								 "x-amz-request-id":"C3D13FE58DE4C810",
								 "x-amz-id-2":"FMyUVURIY8/IgAtTv8xRjskZQpcIZ9KG4V5Wp6S7S/JRWeUWerMUE5JgHvANOjpD"
							},
							"s3":{
								 "s3SchemaVersion":"1.0",
								 "configurationId":"testConfigRule",
								 "bucket":{
										"name":"` + sourceBucketName + `",
										"ownerIdentity":{
											 "principalId":"A3NL1KOZZKExample"
										},
										"arn":"arn:aws:s3:::` + sourceBucketName + `"
								 },
								 "object":{
										"key":"HappyFace.jpg",
										"size":1024,
										"eTag":"d41d8cd98f00b204e9800998ecf8427e",
										"versionId":"096fKKXTRTtl3on89fVO.nfljtsv6qko"
								 }
							}
					 }
				 ]
			 }`),
		})

		if err != nil {
			log.Fatalf("Error from Invoke: %s", err)
		} else {
			logText, err := base64.StdEncoding.DecodeString(*output.LogResult)
			if err != nil {
				log.Fatalf("Couldn't decode Base64: %s", output.LogResult)
			}
			future <- InvokeFunctionReturn{logText: string(logText)}
		}
	}()
	return future
}
