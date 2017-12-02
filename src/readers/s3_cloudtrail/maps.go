package s3_cloudtrail

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
)

type File struct {
	Records []Event
}

type Event struct {
	EventVersion       string                 `json:"eventVersion"`
	UserIdentity       UserIdentity           `json:"userIdentity"`
	EventTime          string                 `json:"eventTime"`
	EventSource        string                 `json:"eventSource"`
	EventName          string                 `json:"eventName"`
	AwsRegion          string                 `json:"awsRegion"`
	SourceIpAddress    string                 `json:"sourceIPAddress"`
	UserAgent          string                 `json:"userAgent"`
	RequestParameters  map[string]interface{} `json:"requestParameters"`
	ResponseElements   map[string]interface{} `json:"responseElements"`
	RequestId          string                 `json:"requestID"`
	EventId            string                 `json:"eventID"`
	EventType          string                 `json:"eventType"`
	RecipientAccountId string                 `json:"recipientAccountId"`
}

type UserIdentity struct {
	Type           string         `json:"type"`
	PrincipalId    string         `json:"principalId"`
	Arn            string         `json:"arn"`
	AccountId      string         `json:"accountId"`
	AccessKeyId    string         `json:"accessKeyId"`
	SessionContext SessionContext `json:"sessionContext"`
}

type SessionContext struct {
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	MfaAuthenticated string `json:"mfaAuthenticated"`
	CreationDate     string `json:"creationDate"`
}

func readJsonIntoEventMaps(reader io.Reader) []map[string]interface{} {
	fileJson, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalw("Error from ReadAll", "err", err)
	}

	file := File{}
	err = json.Unmarshal(fileJson, &file)
	if err != nil {
		log.Fatalw("Error from Unmarshal", "err", err)
	}

	maps := []map[string]interface{}{}
	for _, record := range file.Records {
		timestamp, err := time.Parse(time.RFC3339, record.EventTime)
		if err != nil {
			log.Fatalw("Can't parse timestamp", "eventTime", timestamp)
		}

		m := map[string]interface{}{
			"timestamp":                            timestamp,
			"eventVersion":                         record.EventVersion,
			"eventSource":                          record.EventSource,
			"eventName":                            record.EventName,
			"awsRegion":                            record.AwsRegion,
			"sourceIPAddress":                      record.SourceIpAddress,
			"userAgent":                            record.UserAgent,
			"requestParameters":                    fmt.Sprintf("%+v", record.RequestParameters),
			"responseElements":                     fmt.Sprintf("%+v", record.ResponseElements),
			"requestID":                            record.RequestId,
			"eventID":                              record.EventId,
			"eventType":                            record.EventType,
			"recipientAccountId":                   record.RecipientAccountId,
			"userIdentityType":                     record.UserIdentity.Type,
			"userIdentityPrincipalId":              record.UserIdentity.PrincipalId,
			"userIdentityArn":                      record.UserIdentity.Arn,
			"userIdentityAccountId":                record.UserIdentity.AccountId,
			"userIdentityAccessKeyId":              record.UserIdentity.AccessKeyId,
			"userIdentitySessionContextAttributes": fmt.Sprintf("%+v", record.UserIdentity.SessionContext.Attributes),
		}
		maps = append(maps, m)
	}

	return maps
}
