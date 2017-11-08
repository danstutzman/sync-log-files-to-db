package bigquery

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/cenkalti/backoff"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

type BigqueryConnection struct {
	projectId   string
	datasetName string
	service     *bigquery.Service
}

func (conn *BigqueryConnection) DatasetName() string {
	return conn.datasetName
}

func Atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func ParseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return f
}

func NewBigqueryConnection(opts *Options) *BigqueryConnection {
	pemKeyBytes, err := ioutil.ReadFile(opts.GcloudPemPath)
	if err != nil {
		log.Fatalf("Error from ioutil.ReadFile: %s", err)
	}

	token, err := google.JWTConfigFromJSON(pemKeyBytes, bigquery.BigqueryScope)
	if err != nil {
		log.Fatalf("Error from google.JWTConfigFromJSON: %s", err)
	}
	client := token.Client(oauth2.NoContext)

	service, err := bigquery.New(client)
	if err != nil {
		log.Fatalf("Error from bigquery.New: %s", err)
	}

	return &BigqueryConnection{
		projectId:   opts.GcloudProjectId,
		datasetName: opts.DatasetName,
		service:     service,
	}
}

func (conn *BigqueryConnection) Query(sql, description string) []*bigquery.TableRow {
	var response *bigquery.QueryResponse
	var err error
	err = backoff.Retry(func() error {
		log.Printf("Querying %s...", description)
		response, err = conn.service.Jobs.Query(conn.projectId, &bigquery.QueryRequest{
			Query:        sql,
			UseLegacySql: googleapi.Bool(false),
		}).Do()
		if err != nil {
			err2, isGoogleApiError := err.(*googleapi.Error)
			if isGoogleApiError && (err2.Code == 500 || err2.Code == 503) {
				log.Printf("Got intermittent error (backoff will retry): %s", err2)
				return err
			} else {
				return backoff.Permanent(err)
			}
		} else {
			return nil // success
		}
	}, backoff.NewExponentialBackOff())
	if err != nil {
		log.Fatalf("Error %s; query was %s", err, sql)
	}

	return response.Rows
}

func (conn *BigqueryConnection) CreateDataset() {
	log.Printf("Creating %s dataset...", conn.datasetName)
	_, err := conn.service.Datasets.Insert(conn.projectId, &bigquery.Dataset{
		DatasetReference: &bigquery.DatasetReference{
			ProjectId: conn.projectId,
			DatasetId: conn.datasetName,
		},
	}).Do()

	if err != nil {
		if err.Error() == fmt.Sprintf(
			"googleapi: Error 409: Already Exists: Dataset %s:%s, duplicate",
			conn.projectId, conn.datasetName) {
			// Ignore error
		} else {
			panic(err)
		}
	}
}

func (conn *BigqueryConnection) CreateTable(tableName string,
	fields []*bigquery.TableFieldSchema) {

	log.Printf("Creating %s table...", tableName)
	_, err := conn.service.Tables.Insert(conn.projectId, conn.datasetName,
		&bigquery.Table{
			Schema: &bigquery.TableSchema{Fields: fields},
			TableReference: &bigquery.TableReference{
				DatasetId: conn.datasetName,
				ProjectId: conn.projectId,
				TableId:   tableName,
			},
		}).Do()

	if err != nil {
		if err.Error() == fmt.Sprintf(
			"googleapi: Error 409: Already Exists: Table %s:%s.%s, duplicate",
			conn.projectId, conn.datasetName, tableName) {
			// Ignore error
		} else {
			panic(err)
		}
	}
}

func (conn *BigqueryConnection) InsertRows(tableName string,
	createDataset func(), createTable func(), rows []*bigquery.TableDataInsertAllRequestRows) {

	var err error
	err = backoff.Retry(func() error {
		log.Printf("Inserting rows to %s...", tableName)
		result, err := conn.service.Tabledata.InsertAll(conn.projectId, conn.datasetName,
			tableName, &bigquery.TableDataInsertAllRequest{Rows: rows}).Do()
		if err != nil {
			err2, isGoogleApiError := err.(*googleapi.Error)
			if isGoogleApiError && (err2.Code == 500 || err2.Code == 503) {
				log.Printf("Got intermittent error (backoff will retry): %s", err2)
				return err // Retry with backoff
			} else {
				return backoff.Permanent(err) // Stop backoff
			}
		}

		if len(result.InsertErrors) > 0 {
			for _, errorGroup := range result.InsertErrors {
				for _, e := range errorGroup.Errors {
					log.Printf("InsertError: %v, %v", e.Message, e.Reason)
				}
			}
			return backoff.Permanent(fmt.Errorf("Got InsertErrors"))
		}

		return nil // Success
	}, backoff.NewExponentialBackOff())

	if err != nil {
		if err.Error() == fmt.Sprintf(
			"googleapi: Error 404: Not found: Dataset %s:%s, notFound",
			conn.projectId, conn.datasetName) {

			createDataset()
			panic(fmt.Errorf("Created dataset; maybe wait before restarting"))
		} else if err.Error() == fmt.Sprintf(
			"googleapi: Error 404: Not found: Table %s:%s.%s, notFound",
			conn.projectId, conn.datasetName, tableName) {

			createTable()
			panic(fmt.Errorf("Created table; maybe wait before restarting"))
		} else {
			panic(err)
		}
	}
}
