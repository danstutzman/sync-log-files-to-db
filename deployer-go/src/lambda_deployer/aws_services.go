package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

type LambdaDeployer struct {
	awsSession *session.Session
	s3Service  *s3.S3
	iamService *iam.IAM
}

func NewLambdaDeployer() *LambdaDeployer {
	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &LambdaDeployer{
		awsSession: awsSession,
		s3Service:  s3.New(awsSession),
		iamService: iam.New(awsSession),
	}
}

type CreateRoleIdempotentReturn struct {
	arn string
}

func (lambdaDeployer *LambdaDeployer) CreateRoleIdempotent(roleName string) chan CreateRoleIdempotentReturn {
	return createRoleIdempotent(lambdaDeployer.iamService, roleName)
}

type PutRolePolicyReturn struct{}

func (lambdaDeployer *LambdaDeployer) PutRolePolicy(roleName string, sourceBucket string, targetBucket string) chan PutRolePolicyReturn {
	return putRolePolicy(lambdaDeployer.iamService, roleName, sourceBucket,
		targetBucket)
}

type CreateBucketReturn struct{}

func (lambdaDeployer *LambdaDeployer) CreateBucket(bucketName string) chan CreateBucketReturn {
	return createBucket(lambdaDeployer.s3Service, bucketName)
}

type CopyToBucketReturn struct{}

func (lambdaDeployer *LambdaDeployer) CopyToBucket(fromPath string, toBucketName string, toPath string) chan CopyToBucketReturn {
	return copyToBucket(lambdaDeployer.s3Service, fromPath, toBucketName, toPath)
}
