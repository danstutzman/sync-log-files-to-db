package aws_services

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AwsServices struct {
	awsSession *session.Session
	s3Service  *s3.S3
	iamService *iam.IAM
}

func NewAwsServices() *AwsServices {
	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &AwsServices{
		awsSession: awsSession,
		s3Service:  s3.New(awsSession),
		iamService: iam.New(awsSession),
	}
}
