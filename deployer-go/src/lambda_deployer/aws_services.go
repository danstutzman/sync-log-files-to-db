package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Config struct {
	SourceBucketName string
	TargetBucketName string
	FunctionName     string
	RoleName         string
}

type LambdaDeployer struct {
	config     Config
	awsSession *session.Session
	s3Service  *s3.S3
	iamService *iam.IAM
}

func NewLambdaDeployer(config Config) *LambdaDeployer {
	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &LambdaDeployer{
		config:     config,
		awsSession: awsSession,
		s3Service:  s3.New(awsSession),
		iamService: iam.New(awsSession),
	}
}

func (self *LambdaDeployer) SetupBuckets() {
	createSourceBucketFuture := createBucket(self.s3Service,
		self.config.SourceBucketName)
	createTargetBucketFuture := createBucket(self.s3Service,
		self.config.TargetBucketName)

	<-createSourceBucketFuture
	copySampleImageFuture := copyToBucket(self.s3Service, "../HappyFace.jpg",
		self.config.SourceBucketName, "/HappyFace.jpg")

	<-createTargetBucketFuture
	<-copySampleImageFuture
}

func (self *LambdaDeployer) DeployFunction() {
	roleArn := <-createRoleIdempotent(self.iamService, self.config.RoleName)
	_ = roleArn
	<-putRolePolicy(self.iamService, self.config.RoleName,
		self.config.SourceBucketName, self.config.TargetBucketName)
}

type CreateRoleIdempotentReturn struct {
	arn string
}

type EmptyReturn struct{}
