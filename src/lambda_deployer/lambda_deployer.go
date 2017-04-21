package lambda_deployer

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/src/storage/bigquery"
	my_s3 "github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/src/storage/s3"
	"io/ioutil"
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

func (self *LambdaDeployer) InstallFunction() {
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

func (self *LambdaDeployer) RemoveFunction() {
	//future := deleteBucket(self.s3Service, self.config.BucketName)

	<-deleteFunction(self.lambdaService, self.config.FunctionName)
	<-deleteRolePolicy(self.iamService, self.config.RoleName,
		self.config.PolicyName)
	<-deleteRole(self.iamService, self.config.RoleName)

	//<-future
}

func (self *LambdaDeployer) SyncExistingLogs() {
	bigqueryOptionsBytes, err := ioutil.ReadFile("config/bigquery.json")
	if err != nil {
		log.Fatalf("Error from ReadFile: %s", err)
	}
	var bigqueryOptions bigquery.Options
	json.Unmarshal(bigqueryOptionsBytes, &bigqueryOptions)
	bigquery.ValidateOptions(&bigqueryOptions)
	bigqueryConn := bigquery.NewBigqueryConnection(&bigqueryOptions)

	s3OptionsBytes, err := ioutil.ReadFile("config/s3.json")
	if err != nil {
		log.Fatalf("Error from ReadFile: %s", err)
	}
	var s3Options my_s3.Options
	json.Unmarshal(s3OptionsBytes, &s3Options)
	my_s3.ValidateOptions(&s3Options)
	s3Connection := my_s3.NewS3Connection(&s3Options)

	for _, path := range s3Connection.ListPaths() {
		visits := s3Connection.DownloadVisitsForPath(path)
		bigqueryConn.UploadVisits(path, visits)
		s3Connection.DeletePath(path)
	}
}
