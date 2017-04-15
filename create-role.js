// @flow
const AWS     = require('aws-sdk')
const fs      = require('fs')
const Promise = require('bluebird')

AWS.config.loadFromPath('./config.json');

// Returns Promise with ARN
function createIamRoleIdempotent(roleName:string) {
  return new Promise(function(resolve, reject) {
    getArnForIamRole(roleName).then(function(arn) {
      resolve(arn)
    }).catch(function(err) {
      if (err.code === 'NoSuchEntity') {
        resolve(createIamRoleNonIdempotent(roleName))
      } else {
        reject(err)
      }
    })
  })
}

function getArnForIamRole(roleName:string) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting IAM.getRole for role name '${roleName}'...`)
    new AWS.IAM().getRole({
      RoleName: roleName,
    }, function(err, data) {
      if (err) {
        reject(err)
      } else {
        if (data && data['Role']) {
          const arn = data['Role']['Arn']
          resolve(arn)
        } else {
          reject(`Bad data from getRole: ${JSON.stringify(data)}`)
        }
      }
    })
  })
}

// Returns Promise with ARN
function createIamRoleNonIdempotent(roleName:string) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting IAM.createRole for role name '${roleName}'...`)
    new AWS.IAM().createRole({
      AssumeRolePolicyDocument: JSON.stringify({
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
      }),
      Path: "/",
      RoleName: roleName,
    }, function(err, data) {
      if (err) {
        reject(err)
      } else {
        resolve(getArnForIamRole(roleName))
      }
    })
  })
}

// Returns promise with no data
function putRolePolicyIdempotent(roleName:string,
    lambdaExecutionAccessPolicyName:string, sourceBucket:string, targetBucket:string) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting IAM.putRolePolicy for role name '${roleName}'...`)
    new AWS.IAM().putRolePolicy({
      RoleName: roleName,
      PolicyName: lambdaExecutionAccessPolicyName,
      PolicyDocument: JSON.stringify({
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
            "Resource": `arn:aws:s3:::${sourceBucket}/*`
          },
          {
            "Effect": "Allow",
            "Action": [
              "s3:PutObject"
            ],
            "Resource": `arn:aws:s3:::${targetBucket}/*`
          }
        ]
      })
    }, function(err, data) {
      if (err) {
        reject(err)
      } else {
        resolve()
      }
    })
  })
}


// TODO: remove dependency on CreateThumbnail.zip
function createFunction(functionName:string, executionRoleArn:string) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting Lambda.createFunction for name '${functionName}'...`)
    new AWS.Lambda().createFunction({
      FunctionName: functionName,
      Role: executionRoleArn,
      Timeout: 30,
      Runtime: 'nodejs4.3',
      Code: {
        ZipFile: fs.readFileSync('CreateThumbnail.zip'),
      },
      Handler: `${FUNCTION_NAME}.handler`,
    }, function(err, data) {
      if (err) {
        reject(err)
      } else {
        if (data && data['FunctionArn']) {
          resolve(data['FunctionArn'])
        } else {
          throw new Error(`Couldn't find functionArn in result from createFunction: ${
            JSON.stringify(data)}`)
        }
      }
    })
  })
}

// Returns Promise with functionArn as data
function createFunctionIdempotent(functionName:string, executionRoleArn:string) {
  return new Promise(function(resolve, reject) {
    console.log(
      `Requesting Lambda.listVersionsByFunction for name '${functionName}'...`)
    new AWS.Lambda().listVersionsByFunction({
      FunctionName: functionName,
    }, function(err, data) {
      if (err) {
        if (err.code === 'ResourceNotFoundException') {
          resolve(createFunction(functionName, executionRoleArn))
        } else {
          reject(err)
        }
      } else {
        if (!data || !data.Versions) {
          reject(`Couldn't find Versions in: ${JSON.stringify(data)}`)
        } else {
          let functionArn;
          for (const version of (data.Versions:any)) {
            if (version.Version === '$LATEST') {
              functionArn = version['FunctionArn']
            }
          }
          if (functionArn) {
            resolve(functionArn)
          } else {
            reject(
              `Couldn't find functionArn in versions: ${JSON.stringify(data)}`)
          }
        }
      }
    })
  })
}

function invokeFunction(functionName:string, sourceBucket:string) {
  return new Promise(function(resolve, reject) {
    console.log(
      `Requesting Lambda.invokeFunction for name '${functionName}'...`)
    new AWS.Lambda().invoke({
      FunctionName: functionName,
      InvocationType: 'RequestResponse',
      LogType: 'Tail',
      Payload: JSON.stringify({
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
                   "name":sourceBucket,
                   "ownerIdentity":{
                      "principalId":"A3NL1KOZZKExample"
                   },
                   "arn":`arn:aws:s3:::${sourceBucket}`
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
      }),
    }, function(err, data) {
      if (err) {
        reject(err)
      } else {
        if (data && data['LogResult']) {
          const base64LogText = data['LogResult']
          resolve(Buffer.from(base64LogText, 'base64').toString())
        } else {
          reject(`Unexpected response from invokeFunction: ${JSON.stringify(data)}`)
        }
      }
    })
  })
}

function putBucketNotification(sourceBucket:string, functionArn:string) {
  const functionArnWithoutVersion = functionArn.split(':').slice(0, -1).join(':')
  return new Promise(function(resolve, reject) {
    console.log(`Requesting Lambda.putBucketNotification...`)
    new AWS.S3().putBucketNotification({
      Bucket: sourceBucket,
      NotificationConfiguration: {
        CloudFunctionConfiguration: {
          Event: "s3:ObjectCreated:*",
          CloudFunction: functionArnWithoutVersion,
          Id: "CreateThumbnailStartingEvent",
        }
      },
    }, function(err, data) {
      if (err) {
        reject(JSON.stringify(err))
      } else {
        resolve()
      }
    })
  })
}

function addPermission(functionName:string, sourceBucket:string) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting Lambda.addPermission...`)
    new AWS.Lambda().addPermission({
      FunctionName: functionName,
      Action: 'lambda:InvokeFunction',
      Principal: 's3.amazonaws.com',
      StatementId: 'some-unique-id',
      SourceArn: `arn:aws:s3:::${sourceBucket}`,
    }, function(err, data) {
      if (err) {
        if (err.code === 'ResourceConflictException') { // already exists, so ignore
          resolve()
        } else {
          reject(err)
        }
      } else {
        resolve()
      }
    })
  })
}

const SOURCE_BUCKET = 'danstutzman-lambda-example'
const TARGET_BUCKET = `${SOURCE_BUCKET}resized`
const FUNCTION_NAME = 'CreateThumbnail'
const EXECUTION_ROLE_NAME = `lambda-${FUNCTION_NAME}-execution`
const EXECUTION_POLICY_NAME = `lambda-${FUNCTION_NAME}-execution-access`

createIamRoleIdempotent(EXECUTION_ROLE_NAME).then(function(executionRoleArn) {
  putRolePolicyIdempotent(EXECUTION_ROLE_NAME, EXECUTION_POLICY_NAME, SOURCE_BUCKET,
      TARGET_BUCKET).then(function() {
    createFunctionIdempotent(FUNCTION_NAME, executionRoleArn)
        .then(function(functionArn) {
      addPermission(FUNCTION_NAME, SOURCE_BUCKET).then(function() {
        putBucketNotification(SOURCE_BUCKET, functionArn).then(function() {
          console.log('put bucket notification')
        })
      })
      invokeFunction(FUNCTION_NAME, SOURCE_BUCKET).then(function(logText) {
        console.log('invoke', logText)
      })
    })
  })
}).catch(function(err) {
  console.error('Error', err)
})
