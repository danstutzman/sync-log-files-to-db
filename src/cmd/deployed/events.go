package main

import (
	"encoding/json"
	"strings"
	"log"
)

type EventsJson struct {
	Records []EventRecordJson
}

type EventRecordJson struct {
	EventVersion      string `json:"eventVersion"`
	EventSource       string `json:"eventSource"`
	AwsRegion         string `json:"awsRegion"`
	EventTime         string `json:"eventTime"`
	EventName         string `json:"eventName"`
	UserIdentity      UserIdentityJson `json:"userIdentity"`
	RequestParameters map[string]string `json:"requestParameters"`
	ResponseElements  map[string]string `json:"responseElements"`
	S3                S3Json `json:"s3"`
}

type UserIdentityJson struct {
	PrincipalId string `json:"principalId"`
}

type S3Json struct {
	S3SchemaVersion string `json:"s3SchemaVersion"`
	ConfigurationId string `json:"configurationId"`
	Bucket          BucketJson `json:"bucket"`
	Arn             string `json:"arn"`
	Object          ObjectJson `json:"object"`
}

type BucketJson struct {
	Name          string `json:"name"`
	OwnerIdentity UserIdentityJson `json:"ownerIdentity"`
}

type ObjectJson struct {
	Key       string `json:"key"`
	Size      int `json:"size"`
	ETag      string `json:"eTag"`
	VersionId string `json:"versionId"`
}

func decodeEvents(events string) EventsJson {
	decoder := json.NewDecoder(strings.NewReader(events))
	var eventsJson EventsJson
	err := decoder.Decode(&eventsJson)
	if err != nil {
		log.Fatalf("Error from Decode: %s", err)
	}
	return eventsJson
}
