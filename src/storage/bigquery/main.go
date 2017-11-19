package bigquery

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"

	"github.com/cenkalti/backoff"
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

type BigqueryConnection struct {
	projectId   string
	datasetName string
	tableName   string
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

func NewBigqueryConnection(opts *Options, configPath string) *BigqueryConnection {
	var service *bigquery.Service
	var err error
	if opts.Endpoint == "" {
		gcloudPemPath := path.Join(path.Dir(configPath), opts.GcloudPemPath)
		pemKeyBytes, err := ioutil.ReadFile(gcloudPemPath)
		if err != nil {
			log.Fatalw("Error from ioutil.ReadFile", "err", err)
		}

		token, err := google.JWTConfigFromJSON(pemKeyBytes, bigquery.BigqueryScope)
		if err != nil {
			log.Fatalw("Error from google.JWTConfigFromJSON", "err", err)
		}
		client := token.Client(oauth2.NoContext)

		service, err = bigquery.New(client)
		if err != nil {
			log.Fatalw("Error from bigquery.New", "err", err)
		}
	} else {
		client := &http.Client{}
		service, err = bigquery.New(client)
		if err != nil {
			log.Fatalw("Error from bigquery.New", "err", err)
		}
		service.BasePath = opts.Endpoint
		log.Infow("BasePath", "basePath", service.BasePath)
	}

	return &BigqueryConnection{
		projectId:   opts.GcloudProjectId,
		datasetName: opts.DatasetName,
		tableName:   opts.TableName,
		service:     service,
	}
}

func (conn *BigqueryConnection) Query(sql, description string) []*bigquery.TableRow {
	var response *bigquery.QueryResponse
	var err error
	err = backoff.Retry(func() error {
		log.Infow("Querying", "description", description)
		response, err = conn.service.Jobs.Query(conn.projectId, &bigquery.QueryRequest{
			Query:        sql,
			UseLegacySql: googleapi.Bool(false),
		}).Do()
		if err != nil {
			err2, isGoogleApiError := err.(*googleapi.Error)
			if isGoogleApiError && (err2.Code == 500 || err2.Code == 503) {
				log.Infow("Got intermittent error (backoff will retry)", "err", err2)
				return err
			} else {
				return backoff.Permanent(err)
			}
		} else {
			return nil // success
		}
	}, backoff.NewExponentialBackOff())
	if err != nil {
		log.Fatalw("Got error", "err", err, "query", sql)
	}

	return response.Rows
}

func (conn *BigqueryConnection) CreateDataset() {
	log.Infow("Creating dataset...", "datasetName", conn.datasetName)
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

func (conn *BigqueryConnection) CreateTable(fields []*bigquery.TableFieldSchema) {

	log.Infow("Creating table...", "tableName", conn.tableName)
	_, err := conn.service.Tables.Insert(conn.projectId, conn.datasetName,
		&bigquery.Table{
			Schema: &bigquery.TableSchema{Fields: fields},
			TableReference: &bigquery.TableReference{
				DatasetId: conn.datasetName,
				ProjectId: conn.projectId,
				TableId:   conn.tableName,
			},
		}).Do()

	if err != nil {
		if err.Error() == fmt.Sprintf(
			"googleapi: Error 409: Already Exists: Table %s:%s.%s, duplicate",
			conn.projectId, conn.datasetName, conn.tableName) {
			// Ignore error
		} else {
			panic(err)
		}
	}
}

func (conn *BigqueryConnection) InsertRows(
	maps []map[string]interface{}, uniqueColumnName string) {

	var err error
	err = backoff.Retry(func() error {
		rows := make([]*bigquery.TableDataInsertAllRequestRows, 0)
		for _, m := range maps {
			m2 := map[string]bigquery.JsonValue{}
			for key, value := range m {
				m2[key] = value
			}

			row := &bigquery.TableDataInsertAllRequestRows{
				InsertId: m[uniqueColumnName].(string),
				Json:     m2,
			}
			rows = append(rows, row)
		}

		log.Infow("Inserting rows...", "tableName", conn.tableName)
		result, err := conn.service.Tabledata.InsertAll(
			conn.projectId, conn.datasetName, conn.tableName,
			&bigquery.TableDataInsertAllRequest{Rows: rows}).Do()
		if err != nil {
			err2, isGoogleApiError := err.(*googleapi.Error)
			if isGoogleApiError && (err2.Code == 500 || err2.Code == 503) {
				log.Infow("Got intermittent error (backoff will retry)", "err", err2)
				return err // Retry with backoff
			} else {
				return backoff.Permanent(err) // Stop backoff
			}
		}

		if len(result.InsertErrors) > 0 {
			for _, errorGroup := range result.InsertErrors {
				for _, e := range errorGroup.Errors {
					log.Errorw("InsertError", "err", fmt.Sprintf("%+v", e))
				}
			}
			return backoff.Permanent(fmt.Errorf("Got InsertErrors"))
		}

		return nil // Success
	}, backoff.NewExponentialBackOff())

	if err != nil {
		panic(err)
	}
}
