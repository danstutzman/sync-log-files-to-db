#!/bin/bash -ex

source_bucket=danstutzman-lambda-example

# Do not change this. Walkthrough code assumes this name
target_bucket=${source_bucket}resized

function=CreateThumbnail
lambda_execution_role_name=lambda-$function-execution
lambda_execution_access_policy_name=lambda-$function-execution-access
lambda_invocation_role_name=lambda-$function-invocation
lambda_invocation_access_policy_name=lambda-$function-invocation-access
log_group_name=/aws/lambda/$function

if false; then
aws s3 mb s3://$source_bucket
aws s3 mb s3://$target_bucket

wget -q -OHappyFace.jpg \
  https://c3.staticflickr.com/7/6209/6094281702_d4ac7290d3_b.jpg

aws s3 cp HappyFace.jpg s3://$source_bucket/

wget -q -O $function.js \
    http://run.alestic.com/lambda/aws-examples/CreateThumbnail.js

npm install async gm
fi

zip -r $function.zip $function.js node_modules
exit 1

lambda_execution_role_arn=`
  aws iam get-role --role-name 2$lambda_execution_role_name \
    | python -c 'import json, sys; print json.load(sys.stdin)["Role"]["Arn"]'`
echo $lambda_execution_role_arn
exit 1

lambda_execution_role_arn=$(aws iam create-role \
  --role-name "$lambda_execution_role_name" \
  --assume-role-policy-document '{
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
    }' \
  --output text \
  --query 'Role.Arn'
)
echo lambda_execution_role_arn=$lambda_execution_role_arn

aws iam put-role-policy \
  --role-name "$lambda_execution_role_name" \
  --policy-name "$lambda_execution_access_policy_name" \
  --policy-document '{
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
        "Resource": "arn:aws:s3:::'$source_bucket'/*"
      },
      {
        "Effect": "Allow",
        "Action": [
          "s3:PutObject"
        ],
        "Resource": "arn:aws:s3:::'$target_bucket'/*"
      }
    ]
  }'

lambda_execution_role_arn=arn:aws:iam::553826207523:role/lambda-CreateThumbnail-execution
aws lambda create-function \
  --function-name "$function" \
  --zip-file "fileb://$function.zip" \
  --role "$lambda_execution_role_arn" \
  --handler "$function.handler" \
  --timeout 30 \
  --runtime nodejs4.3
#  --mode event \

  cat > $function-data.json <<EOM
  {  
     "Records":[  
        {  
           "eventVersion":"2.0",
           "eventSource":"aws:s3",
           "awsRegion":"us-east-1",
           "eventTime":"1970-01-01T00:00:00.000Z",
           "eventName":"ObjectCreated:Put",
           "userIdentity":{  
              "principalId":"AIDAJDPLRKLG7UEXAMPLE"
           },
           "requestParameters":{  
              "sourceIPAddress":"127.0.0.1"
           },
           "responseElements":{  
              "x-amz-request-id":"C3D13FE58DE4C810",
              "x-amz-id-2":"FMyUVURIY8/IgAtTv8xRjskZQpcIZ9KG4V5Wp6S7S/JRWeUWerMUE5JgHvANOjpD"
           },
           "s3":{  
              "s3SchemaVersion":"1.0",
              "configurationId":"testConfigRule",
              "bucket":{  
                 "name":"$source_bucket",
                 "ownerIdentity":{  
                    "principalId":"A3NL1KOZZKExample"
                 },
                 "arn":"arn:aws:s3:::$source_bucket"
              },
              "object":{  
                 "key":"HappyFace.jpg",
                 "size":1024,
                 "eTag":"d41d8cd98f00b204e9800998ecf8427e",
                 "versionId":"096fKKXTRTtl3on89fVO.nfljtsv6qko"
              }
           }
        }
     ]
  }
EOM

aws lambda invoke-async \
  --function-name "$function" \
  --invoke-args "$function-data.json"


exit 1
#####################################

#lambda_function_arn=$(aws lambda get-function-configuration \
#  --function-name "$function" \
#  --output text \
#  --query 'FunctionArn'
#)
#echo lambda_function_arn=$lambda_function_arn
lambda_function_arn=arn:aws:lambda:us-east-1:553826207523:function:CreateThumbnail

# get the 553826207523 from aws list-roles output ARN of arn:aws:iam::553826207523:role/lambda-CreateThumbnail-execution"
#aws lambda add-permission \
#  --function-name $function \
#  --region us-east-1 \
#  --statement-id some-unique-id \
#  --action "lambda:InvokeFunction" \
#  --principal s3.amazonaws.com \
#  --source-arn arn:aws:s3:::$source_bucket \
#  --source-account 553826207523

#aws lambda get-policy --function-name $function

lambda_invocation_role_arn=arn:aws:iam::553826207523:role/lambda-CreateThumbnail-execution
aws s3api put-bucket-notification \
  --bucket "$source_bucket" \
  --notification-configuration '{
    "CloudFunctionConfiguration": {
      "CloudFunction": "arn:aws:lambda:us-east-1:553826207523:function:CreateThumbnail",
      "Id": "CreateThumbnailStartingEvent",
      "Event": "s3:ObjectCreated:*"
   }}'

#  --notification-configuration '{
#    "CloudFunctionConfiguration": {
#      "CloudFunction": "'$lambda_function_arn'",
#      "InvocationRole": "'$lambda_invocation_role_arn'",
#      "Event": "s3:ObjectCreated:*"
#    }
#  }'
