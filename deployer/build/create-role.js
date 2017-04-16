//      
const AWS      = require('aws-sdk')
const crypto   = require('crypto')
const execSync = require('child_process').execSync
const fs       = require('fs')
const Promise  = require('bluebird')

AWS.config.loadFromPath('./config.json');
const Lambda = new AWS.Lambda({
  apiVersion: '2015-03-31',
})
const S3 = new AWS.S3({
  apiVersion: '2006-03-01',
})

const SOURCE_BUCKET = 'danstutzman-lambda-example'
const TARGET_BUCKET = `${SOURCE_BUCKET}resized`
const FUNCTION_NAME = 'CreateThumbnail'
const EXECUTION_ROLE_NAME = `lambda-${FUNCTION_NAME}-execution`
const EXECUTION_POLICY_NAME = `lambda-${FUNCTION_NAME}-execution-access`
const ROLE_POLICY_WAIT_SECONDS = 8

// Returns Promise with ARN
function createIamRoleIdempotent(roleName       ) {
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

function getArnForIamRole(roleName       ) {
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
function createIamRoleNonIdempotent(roleName       ) {
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
function putRolePolicyIdempotent(roleName       ,
    lambdaExecutionAccessPolicyName       , sourceBucket       , targetBucket       ) {
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
        console.log('Wait %d seconds for role policy to take effect...',
          ROLE_POLICY_WAIT_SECONDS)
        setTimeout(function() {
          resolve()
        }, ROLE_POLICY_WAIT_SECONDS * 1000)
      }
    })
  })
}


// TODO: remove dependency on CreateThumbnail.zip
function createFunction(functionName       , executionRoleArn       , zipPath       ) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting Lambda.createFunction for name '${functionName}' for executionRoleArn ${executionRoleArn}...`)
    Lambda.createFunction({
      FunctionName: functionName,
      Description: zipPath,
      Publish: true,
      Role: executionRoleArn,
      Timeout: 30,
      Runtime: 'nodejs4.3',
      Code: {
        ZipFile: fs.readFileSync(zipPath),
      },
      Handler: `${FUNCTION_NAME}.handler`,
    }, function(err, data) {
      if (err) {
        reject(err)
      } else {
        if (data && data['FunctionArn']) {
          resolve(data['FunctionArn'])
        } else {
          throw new Error(
            `Couldn't find functionArn in result from createFunction: ${
            JSON.stringify(data)}`)
        }
      }
    }) // end createFunction
  }) // end promise
}

// Returns Promise with Base64 of SHA256 of contents of file
function sha256Base64OfPath(path       ) {
  return new Promise(function(resolve, reject) {
    const sha256 = crypto.createHash('sha256')
    fs.createReadStream(path)
      .on('data', function (chunk) {
        sha256.update(chunk);
      })
      .on('end', function () {
        const sha256Base64 = sha256.digest('base64')
        resolve(sha256Base64)
      })
  })
}

// Returns Promise with functionArn as data
function createFunctionIdempotent(functionName       , executionRoleArn       ) {
  return new Promise(function(resolve, reject) {
    const gitSha1 = execSync('git rev-parse --short HEAD').toString().trim()
    const zipPath = `../deployed/build/${gitSha1}.zip`
    if (!fs.existsSync(zipPath)) {
      const zipCommand = `cd ../deployed &&
        npm install &&
        mkdir -p build &&
        cat src/CreateThumbnail.js > CreateThumbnail.js &&
        zip -r -q ../deployer/${zipPath} CreateThumbnail.js node_modules &&
        rm -f CreateThumbnail.js`
// TODO: check porcelain for uncommitted
      console.log(`Executing ${zipCommand}...`)
      console.log(execSync(zipCommand).toString())
    }
    sha256Base64OfPath(zipPath).then(function(zipSha256) {
      console.log(
        `Requesting Lambda.listVersionsByFunction for name '${functionName}'...`)
      Lambda.listVersionsByFunction({
        FunctionName: functionName,
      }, function(err, data) {
        if (err) {
          if (err.code === 'ResourceNotFoundException') {
            resolve(createFunction(functionName, executionRoleArn, zipPath))
          } else {
            reject(err)
          }
        } else {
          if (!data || !data.Versions) {
            reject(`Couldn't find Versions in: ${JSON.stringify(data)}`)
          } else {
            let matchingFunctionArn;
            for (const version of (data.Versions    )) {
              if (version.Version === '$LATEST' && version.CodeSha256 === zipSha256) {
                matchingFunctionArn =
                  version.FunctionArn.split(':').slice(0, 7).join(':')
              }
            }
            if (matchingFunctionArn) {
              console.log(`Found latest version matching ${zipSha256}`)
              resolve(matchingFunctionArn)
            } else {
              console.log(`Couldn't find latest version matching ${zipSha256}`)
              deleteFunction(functionName, false).then(function() {
                resolve(createFunction(functionName, executionRoleArn, zipPath))
              })
            }
          }
        } // end if
      }) // end listVersionsByFunction
    }) // end then
  }) // end Promise
}

function invokeFunction(functionName       , sourceBucket       ) {
  return new Promise(function(resolve, reject) {
    console.log(
      `Requesting Lambda.invokeFunction for name '${functionName}'...`)
    Lambda.invoke({
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

function putBucketNotification(sourceBucket       , functionArn       ) {
  console.log('functionArn', functionArn)
  return new Promise(function(resolve, reject) {
    console.log(`Requesting S3.putBucketNotification...`)
    S3.putBucketNotification({
      Bucket: sourceBucket,
      NotificationConfiguration: {
        CloudFunctionConfiguration: {
          Event: "s3:ObjectCreated:*",
          CloudFunction: functionArn,
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

function addPermission(functionName       , sourceBucket       ) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting Lambda.addPermission...`)
    Lambda.addPermission({
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

function deleteFunction(functionName       , ignoreIfNotExists     ) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting Lambda.deleteFunction for '${functionName}'...`)
    Lambda.deleteFunction({
      FunctionName: functionName,
    }, function(err, data) {
      if (err) {
        if (ignoreIfNotExists && err.code === 'ResourceNotFoundException') {
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

function deleteRole(roleName       , ignoreIfNotExists     ) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting IAM.deleteRole for '${roleName}'...`)
    new AWS.IAM().deleteRole({
      RoleName: roleName,
    }, function(err, data) {
      if (err) {
        if (ignoreIfNotExists && err.code == 'NoSuchEntity') {
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

function deleteRolePolicy(roleName       , policyName       , ignoreIfNotExists     ) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting IAM.deleteRolePolicy for '${policyName}'...`)
    new AWS.IAM().deleteRolePolicy({
      RoleName: roleName,
      PolicyName: policyName,
    }, function(err, data) {
      if (err) {
        if (ignoreIfNotExists && err.code === 'NoSuchEntity') {
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

function deleteS3BucketRecursively(bucketName       , ignoreIfNotExists     ) {
  return new Promise(function(resolve, reject) {
    console.log(`Requesting S3.deleteBucket for '${bucketName}'...`)
    S3.deleteBucket({
      Bucket: bucketName,
    }, function(err, data) {
      if (err) {
        if (ignoreIfNotExists && err.code === 'NoSuchBucket') {
          resolve()
        } else if (err.code === 'BucketNotEmpty') {
          console.log(`Requesting S3.listObjects for '${bucketName}'...`)
          S3.listObjects({
            Bucket: bucketName,
          }, function(err, data) {
            if (err) {
              reject(err)
            } else {
              const objectsToDelete = []
              for (const object of (data    ).Contents) {
                objectsToDelete.push({
                  Key:       (object    ).Key,
                  VersionId: (object    ).VersionId,
                })
              }

              console.log(`Requesting S3.deleteObjects for '${bucketName}'...`)
              S3.deleteObjects({
                Bucket: bucketName,
                Delete: {
                  Objects: objectsToDelete,
                },
              }, function(err, data) {
                if (err) {
                  reject(err)
                } else {
                  resolve(deleteS3BucketRecursively(bucketName, ignoreIfNotExists))
                }
              })
            }
          })
        } else { // if not BucketNotEmpty
          reject(err)
        }
      } else { // if not err
        resolve()
      }
    })
  })
}

if (false) {
  deleteS3BucketRecursively(SOURCE_BUCKET, true).then(function() {
    deleteS3BucketRecursively(TARGET_BUCKET, true).then(function() {
      deleteFunction(FUNCTION_NAME, true).then(function() {
        deleteRolePolicy(EXECUTION_ROLE_NAME, EXECUTION_POLICY_NAME, true)
            .then(function() {
          deleteRole(EXECUTION_ROLE_NAME, true).then(function() {
            console.log('deleted')
          })
        })
      })
    })
  })
}
if (true) {
  const setupCommands = `aws s3 mb s3://${SOURCE_BUCKET} &&
    aws s3 mb s3://${TARGET_BUCKET} &&
    aws s3 cp ../HappyFace.jpg s3://${SOURCE_BUCKET}/HappyFace.jpg`
  console.log(`Running ${setupCommands}...`)
  console.log(execSync(setupCommands).toString())

  createIamRoleIdempotent(EXECUTION_ROLE_NAME).then(function(executionRoleArn) {
    console.log('executionRoleArn', executionRoleArn)
    putRolePolicyIdempotent(EXECUTION_ROLE_NAME, EXECUTION_POLICY_NAME,
        SOURCE_BUCKET, TARGET_BUCKET).then(function() {
      createFunctionIdempotent(FUNCTION_NAME, executionRoleArn)
          .then(function(functionArn) {
        console.log('functionArn', functionArn)
        addPermission(FUNCTION_NAME, SOURCE_BUCKET).then(function() {
          putBucketNotification(SOURCE_BUCKET, functionArn).then(function() {
            console.log('put bucket notification')
            invokeFunction(FUNCTION_NAME, SOURCE_BUCKET).then(function(logText) {
              console.log('invoke', logText)
            })
          })
        })
      })
    })
  }).catch(function(err) {
    console.error('Error', err)
  })
}
