package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"log"
)

func deleteRolePolicy(iamService *iam.IAM, roleName string, policyName string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Deleting policy named %s...", policyName)
		_, err := iamService.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(policyName),
		})

		if err != nil {
			if err2, ok := err.(awserr.Error); ok && err2.Code() ==
				iam.ErrCodeNoSuchEntityException {
				// Ignore
				future <- EmptyReturn{}
			} else {
				log.Fatalf("Error from DeleteRolePolicy: %s", err)
			}
		} else {
			future <- EmptyReturn{}
		}
	}()
	return future
}
