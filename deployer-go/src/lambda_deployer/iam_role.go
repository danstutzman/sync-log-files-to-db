package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"log"
)

func createRoleNonIdempotent(iamService *iam.IAM,
	roleName string) chan EmptyReturn {

	future := make(chan EmptyReturn)
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

		future <- EmptyReturn{}
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

func createRoleIdempotent(iamService *iam.IAM, roleName string) chan CreateRoleIdempotentReturn {
	future := make(chan CreateRoleIdempotentReturn)
	go func() {
		arn, err := getArnForRole(iamService, roleName)
		if err != nil {
			if err.Error() == iam.ErrCodeNoSuchEntityException {
				<-createRoleNonIdempotent(iamService, roleName)
				arn, err = getArnForRole(iamService, roleName)
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
