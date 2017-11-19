package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

var discoveryJson []byte
var DATASETS_REGEXP = regexp.MustCompile("^(/bigquery/v2)?/projects/(.*?)/datasets$")
var TABLES_REGEXP = regexp.MustCompile("^(/bigquery/v2)?/projects/(.*?)/datasets/(.*?)/tables$")
var JOBS_REGEXP = regexp.MustCompile("^(/bigquery/v2)?/projects/(.*?)/jobs$")
var QUERY_REGEXP = regexp.MustCompile("^(/bigquery/v2)?/projects/(.*?)/queries/(.*?)$")
var INSERT_REGEXP = regexp.MustCompile("^(/bigquery/v2)?/projects/(.*?)/datasets/(.*?)/tables/(.*?)/insertAll")

type CreateDatasetRequest struct {
	DatasetReference DatasetReference `json:"datasetReference"`
}

type DatasetReference struct {
	DatasetId string `json:"datasetId"`
	ProjectId string `json:"projectId"`
}

type CreateJobRequest struct {
	Configuration Configuration `json:"configuration"`
	JobReference  JobReference  `json:"jobReference"`
}

type Configuration struct {
	Query1 Query1 `json:"query"`
}

type Query1 struct {
	Query2 string `json:"query"`
}

type JobReference struct {
	ProjectId string `json:"projectId"`
	JobId     string `json:"jobId"`
}

func serveDiscovery(w http.ResponseWriter, r *http.Request, discoveryJson []byte) {
	w.Write(discoveryJson)
}

func listDatasets(w http.ResponseWriter, r *http.Request, project string) {
	dataset := "belugacdn_logs"
	fmt.Fprintf(w, `{
		"kind": "bigquery#datasetList",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/qwnfLrlOKTXd94DjXLYMd9AnLA8\"",
		"datasets": [
		 {
			"kind": "bigquery#dataset",
			"id": "%s:%s",
			"datasetReference": {
			 "datasetId": "%s",
			 "projectId": "%s"
			}
		 }
		]
	 }`, project, dataset, dataset, project)
}

func createDataset(w http.ResponseWriter, r *http.Request, projectName string) {
	decoder := json.NewDecoder(r.Body)
	var body CreateDatasetRequest
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	projectName2 := body.DatasetReference.ProjectId
	if projectName2 != projectName {
		log.Fatalf("Expected project name to match")
	}
	datasetName := body.DatasetReference.DatasetId

	project, projectOk := projects[projectName]
	if !projectOk {
		project = Project{Datasets: map[string]Dataset{}}
		projects[projectName] = project
	}

	project.Datasets[datasetName] = Dataset{Tables: map[string]Table{}}

	// Just serve the input as output
	outputJson, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("Error from Marshal: %v", err)
	}
	w.Write(outputJson)
}

func listTables(w http.ResponseWriter, r *http.Request, project, dataset string) {
	table := "visits"
	fmt.Fprintf(w, `{
		"kind": "bigquery#tableList",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/zZCSENSD7Bu0j7yv3iZTn_ilPBg\"",
		"tables": [
			{
				"kind": "bigquery#table",
				"id": "%s:%s.%s",
				"tableReference": {
					"projectId": "%s",
					"datasetId": "%s",
					"tableId": "%s"
				},
				"type": "TABLE",
				"creationTime": "1510171319097"
			}
		],
		"totalItems": 1
		}
	`, project, dataset, table, project, dataset, table)
}

type CreateTableRequest struct {
	TableReference TableReference `json:"tableReference"`
	Schema         Schema         `json:"schema"`
}

type TableReference struct {
	ProjectId string `json:"projectId"`
	DatasetId string `json:"datasetId"`
	TableId   string `json:"tableId"`
}

type Schema struct {
	Fields []Field `json:"fields"`
}

type Field struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func createTable(w http.ResponseWriter, r *http.Request, projectName, datasetName string) {
	decoder := json.NewDecoder(r.Body)
	var body CreateTableRequest
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	projectName2 := body.TableReference.ProjectId
	if projectName2 != projectName {
		log.Fatalf("Expected project name to match")
	}
	datasetName2 := body.TableReference.DatasetId
	if datasetName2 != datasetName {
		log.Fatalf("Expected dataset name to match")
	}
	tableName := body.TableReference.TableId

	project, projectOk := projects[projectName]
	if !projectOk {
		project = Project{Datasets: map[string]Dataset{}}
		projects[projectName] = project
	}

	dataset, datasetOk := project.Datasets[datasetName]
	if !datasetOk {
		log.Fatalf("Dataset doesn't exist: %s", datasetName)
	}
	dataset.Tables[tableName] = Table{}

	// Just serve the input as output
	outputJson, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("Error from Marshal: %v", err)
	}
	w.Write(outputJson)

}

func createJob(w http.ResponseWriter, r *http.Request, project string) {
	decoder := json.NewDecoder(r.Body)
	var body CreateJobRequest
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	query := body.Configuration.Query1.Query2
	jobId := body.JobReference.JobId
	queryByJobId[jobId] = query

	email := "a@b.com"
	fmt.Fprintf(w, `{
		"kind": "bigquery#job",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/_oiKSu1NLem_L8Icwp_IYkfy3vg\"",
		"id": "%s:%s",
		"selfLink": "https://www.googleapis.com/bigquery/v2/projects/%s/jobs/%s",
		"jobReference": {
		 "projectId": "%s",
		 "jobId": "%s"
		},
		"configuration": {
		 "query": {
			"query": "%s",
			"destinationTable": {
			 "projectId": "%s",
			 "datasetId": "_2cf7cfaa9c05dd2381014b72df999b53fd45fe85",
			 "tableId": "anon5fb7e0264db7f54e07e3df0833fbfcfd11d63e03"
			},
			"createDisposition": "CREATE_IF_NEEDED",
			"writeDisposition": "WRITE_TRUNCATE"
		 }
		},
		"status": {
		 "state": "DONE"
		},
		"statistics": {
		 "creationTime": "1511049825816",
		 "startTime": "1511049826072"
		},
		"user_email": "%s"
	 }`, project, jobId,
		project, jobId,
		project, jobId,
		query,
		project,
		email)
}

func serveQuery(w http.ResponseWriter, r *http.Request, project, jobId string) {
	query := queryByJobId[jobId]

	fmt.Fprintf(w, `{
		"kind": "bigquery#getQueryResultsResponse",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/wLFL5h11OCxiWY3yDLqREwltkXs\"",
		"schema": {
			"fields": [
				{
					"name": "f0_",
					"type": "INTEGER",
					"mode": "NULLABLE"
				}
			]
		},
		"jobReference": {
			"projectId": "%s",
			"jobId": "%s"
		},
		"totalRows": "1",
		"rows": [
			{
				"f": [
					{
						"v": "%s"
					}
				]
			}
		],
		"totalBytesProcessed": "0",
		"jobComplete": true,
		"cacheHit": true
	}`, project, jobId, query)
}

func insertRows(w http.ResponseWriter, r *http.Request, projectName, datasetName, tableName string) {
	decoder := json.NewDecoder(r.Body)
	var row map[string]interface{}
	err := decoder.Decode(&row)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	project, projectOk := projects[projectName]
	if !projectOk {
		project = Project{Datasets: map[string]Dataset{}}
		projects[projectName] = project
	}

	dataset, datasetOk := project.Datasets[datasetName]
	if !datasetOk {
		log.Fatalf("Dataset doesn't exist: %s", datasetName)
	}

	table, tableOk := dataset.Tables[tableName]
	if !tableOk {
		log.Fatalf("Table doesn't exist: %s", tableName)
	}

	table.Rows = append(table.Rows, row)

	// No errors implies success
	fmt.Fprintf(w, `{
		"kind": "bigquery#tableDataInsertAllResponse"
	}`)
}

func serve(w http.ResponseWriter, r *http.Request, discoveryJson []byte) {
	path := r.URL.Path
	log.Printf("Incoming path: %s", path)

	if path == "/discovery/v1/apis/bigquery/v2/rest" {
		serveDiscovery(w, r, discoveryJson)
	} else if match := DATASETS_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[2]
		if r.Method == "GET" {
			listDatasets(w, r, project)
		} else if r.Method == "POST" {
			createDataset(w, r, project)
		} else {
			log.Fatalf("Unexpected method: %s", r.Method)
		}
	} else if match := TABLES_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[2]
		dataset := match[3]
		if r.Method == "GET" {
			listTables(w, r, project, dataset)
		} else if r.Method == "POST" {
			createTable(w, r, project, dataset)
		} else {
			log.Fatalf("Unexpected method: %s", r.Method)
		}
	} else if match := JOBS_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[2]
		if r.Method == "POST" {
			createJob(w, r, project)
		} else {
			log.Fatalf("Unexpected method: %s", r.Method)
		}
	} else if match := QUERY_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[2]
		jobId := match[3]
		serveQuery(w, r, project, jobId)
	} else if match := INSERT_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[2]
		dataset := match[3]
		table := match[4]
		insertRows(w, r, project, dataset, table)
	} else {
		log.Fatalf("Don't know how to serve path %s", r.URL.Path)
	}
}

func listenAndServe(discoveryJson []byte, portNum int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, discoveryJson)
	})

	log.Printf("Listening on :%d...", portNum)
	err := http.ListenAndServe(fmt.Sprintf(":%d", portNum), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
