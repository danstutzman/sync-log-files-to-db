package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"log"
)

func deleteRole(iamService *iam.IAM, roleName string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Deleting role named %s...", roleName)
		_, err := iamService.DeleteRole(&iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})

		if err != nil {
			log.Fatalf("Error from DeleteRole: %s", err)
		} else {
			future <- EmptyReturn{}
		}
	}()
	return future
}
