package aws_services

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
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

type CreateRoleNonIdempotentReturn struct{}

func createRoleNonIdempotent(iamService *iam.IAM,
	roleName string) chan CreateRoleNonIdempotentReturn {

	future := make(chan CreateRoleNonIdempotentReturn)
	go func() {
		log.Printf("Creating role named %s...", roleName)

		output, err := iamService.CreateRole(&iam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Sid": "",
						"Effect": "Allow",
						"Principal": {
							"Service": "lambda.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					}
				]
			}`),
			Path:     aws.String("/"),
			RoleName: aws.String(roleName),
		})
		if err != nil {
			log.Fatalf("Error from CreateRole: %s", err)
		} else {
			log.Printf("Output from CreateRole: %s", output)
		}

		future <- CreateRoleNonIdempotentReturn{}
	}()
	return future
}

func getArnForRole(iamService *iam.IAM, roleName string) (string, error) {
	output, err := iamService.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return "", err
	} else {
		log.Printf("Output from GetRole: %s", output)
		return *output.Role.Arn, err
	}
}

type CreateRoleIdempotentReturn struct {
	arn string
}

func (awsServices *AwsServices) CreateRoleIdempotent(roleName string) chan CreateRoleIdempotentReturn {
	future := make(chan CreateRoleIdempotentReturn)
	go func() {
		arn, err := getArnForRole(awsServices.iamService, roleName)
		if err != nil {
			if err.Error() == iam.ErrCodeNoSuchEntityException {
				<-createRoleNonIdempotent(awsServices.iamService, roleName)
				arn, err = getArnForRole(awsServices.iamService, roleName)
				if err != nil {
					log.Fatalf("Error from second GetRole: %s", err)
				}
			} else {
				log.Fatalf("Error from first GetRole: %s", err)
			}
		}
		future <- CreateRoleIdempotentReturn{arn: arn}
	}()
	return future
}
