package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

type Config struct {
	SourceBucketName string
	TargetBucketName string
	FunctionName     string
	RoleName         string
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
	roleArn :=
		(<-createRoleIdempotent(self.iamService, self.config.RoleName)).roleArn
	<-putRolePolicy(self.iamService, self.config.RoleName,
		self.config.SourceBucketName, self.config.TargetBucketName)

	zipPath := zip("../deployed")
	log.Printf("sha256base64 %s", sha256Base64(zipPath))
	functionArn := (<-uploadZip(self.lambdaService, zipPath,
		self.config.FunctionName, roleArn)).functionArn
	<-addPermission(self.lambdaService, self.config.FunctionName,
		self.config.SourceBucketName)
	<-putBucketNotification(self.s3Service, self.config.SourceBucketName,
		functionArn)
	logText := (<-invokeFunction(self.lambdaService, self.config.FunctionName,
		self.config.SourceBucketName)).logText
	log.Printf("LogText: %s", logText)
}

type CreateRoleIdempotentReturn struct {
	roleArn string
}

type EmptyReturn struct{}

func (self *LambdaDeployer) DeleteEverything() {
	future1 := deleteBucket(self.s3Service, self.config.SourceBucketName)
	future2 := deleteBucket(self.s3Service, self.config.TargetBucketName)
	future3 := deleteFunction(self.lambdaService, self.config.FunctionName)
	<-future1
	<-future2
	<-future3
}
