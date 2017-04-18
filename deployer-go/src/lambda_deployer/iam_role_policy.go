package lambda_deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"log"
)

func putRolePolicy(iamService *iam.IAM, roleName string, sourceBucket string, targetBucket string) chan PutRolePolicyReturn {
	future := make(chan PutRolePolicyReturn)
	go func() {
		log.Printf("Put role policy on %s...", roleName)

		output, err := iamService.PutRolePolicy(&iam.PutRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(roleName + "-policy"),
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
				 			 "s3:GetObject"
				 		],
				 		"Resource": "arn:aws:s3:::` + sourceBucket + `/*"
					},
					{
				 		"Effect": "Allow",
				 		"Action": [
				 			 "s3:PutObject"
				 		],
				 		"Resource": "arn:aws:s3:::` + targetBucket + `/*"
					}
				]
			}`),
		})
		if err != nil {
			log.Fatalf("Error from PutRolePolicy: %s", err)
		} else {
			log.Printf("Output from PutRolePolicy: %s", output)
		}

		future <- PutRolePolicyReturn{}
	}()
	return future
}
