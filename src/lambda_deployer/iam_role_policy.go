package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"log"
)

func putRolePolicy(iamService *iam.IAM, roleName, policyName, bucketName string) chan EmptyReturn {
	future := make(chan EmptyReturn)
	go func() {
		log.Printf("Put role policy on %s...", roleName)

		output, err := iamService.PutRolePolicy(&iam.PutRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(policyName),
			PolicyDocument: aws.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
				 		"Effect": "Allow",
				 		"Action": [
				 			 "logs:*"
				 		],
				 		"Resource": "arn:aws:logs:*:*:*"
					},
					{
				 		"Effect": "Allow",
				 		"Action": [
				 			 "s3:GetObject",
				 			 "s3:DeleteObject"
				 		],
				 		"Resource": "arn:aws:s3:::` + bucketName + `/*"
					},					{
				 		"Effect": "Allow",
				 		"Action": [
				 			 "s3:ListBucket"
				 		],
				 		"Resource": "arn:aws:s3:::` + bucketName + `"
					}
				]
			}`),
		})
		if err != nil {
			log.Fatalf("Error from PutRolePolicy: %s", err)
		} else {
			log.Printf("Output from PutRolePolicy: %s", output)
		}

		future <- EmptyReturn{}
	}()
	return future
}
