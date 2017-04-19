package lambda_deployer

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type UploadZipReturn struct {
	functionArn string
}

func createFunction(lambdaService *lambda.Lambda, zipPath string, functionName string, roleArn string) UploadZipReturn {
	zipBytes, err := ioutil.ReadFile(zipPath)
	if err != nil {
		log.Fatalf("Error from ReadFile of %s: %s", zipPath, err)
	}

	log.Printf("Creating function %s...", functionName)
	output, err := lambdaService.CreateFunction(&lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			ZipFile: zipBytes,
		},
		FunctionName: aws.String(functionName),
		Handler:      aws.String("NodeWrapper.handler"),
		Role:         aws.String(roleArn),
		Runtime:      aws.String("nodejs4.3"),
		Description:  aws.String(zipPath),
		Timeout:      aws.Int64(30),
	})
	if err != nil {
		log.Fatalf("Error from CreateFunction: %s", err)
		return UploadZipReturn{functionArn: ""} // never executed
	} else {
		return UploadZipReturn{functionArn: *output.FunctionArn}
	}
}

func deleteFunctionSync(lambdaService *lambda.Lambda, functionName string) {
	log.Printf("Deleting %s...", functionName)
	_, err := lambdaService.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		log.Printf("Error from DeleteFunction: %s", err)
	}
}

func uploadZip(lambdaService *lambda.Lambda, zipPath string, functionName string, roleArn string) chan UploadZipReturn {
	future := make(chan UploadZipReturn)
	go func() {
		zipSha256 := sha256Base64(zipPath)

		log.Printf("Listing versions of %s...", functionName)
		output, err := lambdaService.ListVersionsByFunction(
			&lambda.ListVersionsByFunctionInput{
				FunctionName: aws.String(functionName),
			})
		if err != nil {
			if err2, ok := err.(awserr.Error); ok && err2.Code() ==
				lambda.ErrCodeResourceNotFoundException {
				future <- createFunction(lambdaService, zipPath, functionName, roleArn)
			} else {
				log.Fatalf("Error from ListVersionsByFunction: %s", err)
			}
		} else {
			log.Printf("Output from ListVersionsByFunction: %s", output)

			var matchingFunctionArn string
			for _, version := range output.Versions {
				if *version.Version == "$LATEST" && *version.CodeSha256 == zipSha256 {
					matchingFunctionArn =
						strings.Join(strings.Split(*version.FunctionArn, ":")[0:7], ":")
					break
				}
			}

			if matchingFunctionArn != "" {
				log.Printf("Found latest version matching %s", zipSha256)
				future <- UploadZipReturn{functionArn: matchingFunctionArn}
			} else {
				log.Printf("Couldn't find latest version matching %s", zipSha256)
				deleteFunctionSync(lambdaService, functionName)
				future <- createFunction(lambdaService, zipPath, functionName, roleArn)
			}
		}
	}()
	return future
}

func sha256Base64(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error from Open: %s", path)
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		log.Fatalf("Error from Copy: %s", err)
	}

	sha256Bytes := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(sha256Bytes)
}
