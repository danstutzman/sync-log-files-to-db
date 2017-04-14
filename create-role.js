const AWS     = require('aws-sdk')
const fs      = require('fs')
const Promise = require('bluebird')

AWS.config.loadFromPath('./config.json');

// Returns Promise with ARN
function createIamRoleIdempotent(roleName) {
  return getArnForIamRole(roleName).catch(function(err) {
    if (err) {
      if (err.code === 'NoSuchEntity') {
        resolve(createIamRoleNonIdempotent(roleName))
      } else {
        reject(err)
      }
    } else {
      resolve(arn)
    }
  })
}

function getArnForIamRole(roleName) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting IAM.getRole for role name '${roleName}'...`)
    new AWS.IAM().getRole({
      RoleName: roleName,
    }, function(err, data) {
      if (err) {
        reject(err)
      } else {
        const arn = data['Role']['Arn']
        resolve(arn)
      }
    })
  })
}

// Returns Promise with ARN
function createIamRoleNonIdempotent(roleName) {
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
function putRolePolicyIdempotent(roleName, lambdaExecutionAccessPolicyName,
    sourceBucket, targetBucket) {
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
function createFunction(functionName, executionRoleArn) {
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
        resolve(data)
      }
    })
  })
}

function createFunctionIdempotent(functionName, executionRoleArn) {
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
        console.log('got data', data)
        resolve(data)
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
    createFunctionIdempotent(FUNCTION_NAME, executionRoleArn).then(function(data) {
      console.log('createdFunction', data)
    })
  })
}).catch(function(err) {
  console.error('Error', err)
})
