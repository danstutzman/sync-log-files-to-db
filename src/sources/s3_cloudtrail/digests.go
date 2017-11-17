package s3_cloudtrail

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

type Digest struct {
	AwsAccountId                string    `json:"awsAccountId"`
	DigestStartTime             string    `json:"digestStartTime"`
	DigestEndTime               string    `json:"digestEndTime"`
	DigestS3Bucket              string    `json:"digestS3Bucket"`
	DigestS3Object              string    `json:"digestS3Object"`
	DigestPublicKeyFingerprint  string    `json:"digestPublicKeyFingerprint"`
	DigestSignatureAlgorithm    string    `json:"digestSignatureAlgorithm"`
	NewestEventTime             string    `json:"newestEventTime"`
	OldestEventTime             string    `json:"oldestEventTime"`
	PreviousDigestS3Bucket      string    `json:"previousDigestS3Bucket"`
	PreviousDigestS3Object      string    `json:"previousDigestS3Object"`
	PreviousDigestHashValue     string    `json:"previousDigestHashValue"`
	PreviousDigestHashAlgorithm string    `json:"previousDigestHashAlgorithm"`
	PreviousDigestSignature     string    `json:"previousDigestSignature"`
	LogFiles                    []LogFile `json:"logFiles"`
}

type LogFile struct {
	S3Bucket        string `json:"s3Bucket"`
	S3Object        string `json:"s3Object"`
	HashValue       string `json:"hashValue"`
	HashAlgorithm   string `json:"hashAlgorithm"`
	NewestEventTime string `json:"newestEventTime"`
	OldestEventTime string `json:"oldestEventTime"`
}

func readJsonIntoDigestMap(reader io.Reader) map[string]interface{} {
	digestJson, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("Error from ReadAll: %s", err))
	}

	digest := Digest{}
	err = json.Unmarshal(digestJson, &digest)
	if err != nil {
		panic(fmt.Errorf("Error from json.Unmarshal: %s", err))
	}

	if len(digest.LogFiles) > 0 {
		log.Fatalf("Found some LogFiles: %v", digest.LogFiles)
	}

	m := map[string]interface{}{
		"awsAccountId":                digest.AwsAccountId,
		"digestStartTime":             digest.DigestStartTime,
		"digestEndTime":               digest.DigestEndTime,
		"digestS3Bucket":              digest.DigestS3Bucket,
		"digestS3Object":              digest.DigestS3Object,
		"digestPublicKeyFingerprint":  digest.DigestPublicKeyFingerprint,
		"digestSignatureAlgorithm":    digest.DigestSignatureAlgorithm,
		"newestEventTime":             digest.NewestEventTime,
		"oldestEventTime":             digest.OldestEventTime,
		"previousDigestS3Bucket":      digest.PreviousDigestS3Bucket,
		"previousDigestS3Object":      digest.PreviousDigestS3Object,
		"previousDigestHashValue":     digest.PreviousDigestHashValue,
		"previousDigestHashAlgorithm": digest.PreviousDigestHashAlgorithm,
		"previousDigestSignature":     digest.PreviousDigestSignature,
		"logFiles":                    fmt.Sprintf("%v", digest.LogFiles),
	}

	return m
}
