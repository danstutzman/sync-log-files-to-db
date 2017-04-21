package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

type Config struct {
	BucketName   string
	FunctionName string
	RoleName     string
	PolicyName   string
}

type LambdaDeployer struct {
	config        Config
	awsSession    *session.Session
	s3Service     *s3.S3
	iamService    *iam.IAM
	lambdaService *lambda.Lambda
}

func NewLambdaDeployer(config Config) *LambdaDeployer {
	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &LambdaDeployer{
		config:        config,
		awsSession:    awsSession,
		s3Service:     s3.New(awsSession),
		iamService:    iam.New(awsSession),
		lambdaService: lambda.New(awsSession),
	}
}

func (self *LambdaDeployer) SetupBucket() {
	<-createBucket(self.s3Service, self.config.BucketName)
}

func (self *LambdaDeployer) DeployFunction() {
	roleArn :=
		(<-createRoleIdempotent(self.iamService, self.config.RoleName)).roleArn
	<-putRolePolicy(self.iamService, self.config.RoleName, self.config.PolicyName,
		self.config.BucketName)

	zipPath := zip()
	log.Printf("sha256base64 %s", sha256Base64(zipPath))
	functionArn := (<-uploadZip(self.lambdaService, zipPath,
		self.config.FunctionName, roleArn)).functionArn
	<-addPermission(self.lambdaService, self.config.FunctionName,
		self.config.BucketName)
	<-putBucketNotification(self.s3Service, self.config.BucketName,
		functionArn)
	logText := (<-invokeFunction(self.lambdaService, self.config.FunctionName,
		self.config.BucketName)).logText
	log.Printf("LogText: %s", logText)
}

type CreateRoleIdempotentReturn struct {
	roleArn string
}

type EmptyReturn struct{}

func (self *LambdaDeployer) DeleteEverything() {
	//future := deleteBucket(self.s3Service, self.config.BucketName)

	<-deleteFunction(self.lambdaService, self.config.FunctionName)
	<-deleteRolePolicy(self.iamService, self.config.RoleName,
		self.config.PolicyName)
	<-deleteRole(self.iamService, self.config.RoleName)

	//<-future
}
