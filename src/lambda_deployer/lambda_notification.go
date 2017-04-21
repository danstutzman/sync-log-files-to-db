package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func putBucketNotification(s3Service *s3.S3, bucketName string, functionArn string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Putting bucket notification for %s...", bucketName)
		output, err := s3Service.PutBucketNotification(
			&s3.PutBucketNotificationInput{
				Bucket: aws.String(bucketName),
				NotificationConfiguration: &s3.NotificationConfigurationDeprecated{
					CloudFunctionConfiguration: &s3.CloudFunctionConfiguration{
						CloudFunction: aws.String(functionArn),
						Event:         aws.String("s3:ObjectCreated:*"),
						Id:            aws.String("SyncCloudfrontLogsToBigqueryTestEvent"),
					},
				},
			})

		if err != nil {
			log.Fatalf("Error from PutBucketNotification: %s", err)
		} else {
			log.Printf("Output from PutBucketNotification: %v", output)
			future <- EmptyReturn{}
		}
	}()
	return future
}
