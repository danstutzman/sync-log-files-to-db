package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var discoveryJson []byte

func serveDiscovery(w http.ResponseWriter, r *http.Request) {
	w.Write(discoveryJson)
}

func serveDatasets(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{
		"kind": "bigquery#datasetList",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/qwnfLrlOKTXd94DjXLYMd9AnLA8\"",
		"datasets": [
		 {
			"kind": "bigquery#dataset",
			"id": "speech-danstutzman:belugacdn_logs",
			"datasetReference": {
			 "datasetId": "belugacdn_logs",
			 "projectId": "speech-danstutzman"
			}
		 }
		]
	 }`)
}

func serveTables(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{
		"kind": "bigquery#tableList",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/zZCSENSD7Bu0j7yv3iZTn_ilPBg\"",
		"tables": [
			{
			"kind": "bigquery#table",
			"id": "speech-danstutzman:belugacdn_logs.visits",
			"tableReference": {
				"projectId": "speech-danstutzman",
				"datasetId": "belugacdn_logs",
				"tableId": "visits"
			},
			"type": "TABLE",
			"creationTime": "1510171319097"
			}
		],
		"totalItems": 1
		}
	`)
}

func startJob(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{
		"kind": "bigquery#job",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/_oiKSu1NLem_L8Icwp_IYkfy3vg\"",
		"id": "speech-danstutzman:bqjob_r7c51234c0123569f_0000015fd1968828_1",
		"selfLink": "https://www.googleapis.com/bigquery/v2/projects/speech-danstutzman/jobs/bqjob_r7c51234c0123569f_0000015fd1968828_1",
		"jobReference": {
		 "projectId": "speech-danstutzman",
		 "jobId": "bqjob_r7c51234c0123569f_0000015fd1968828_1"
		},
		"configuration": {
		 "query": {
			"query": "select count(*) from belugacdn_logs.visits",
			"destinationTable": {
			 "projectId": "speech-danstutzman",
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
		"user_email": "dtstutz@gmail.com"
	 }`)
}

func serveQuery(w http.ResponseWriter, r *http.Request) {
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
			"projectId": "speech-danstutzman",
			"jobId": "bqjob_r6c744039b40f818e_0000015fd19a3130_1"
		},
		"totalRows": "1",
		"rows": [
			{
			"f": [
				{
				"v": "704"
				}
			]
			}
		],
		"totalBytesProcessed": "0",
		"jobComplete": true,
		"cacheHit": true
	}`)
}

func serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/discovery/v1/apis/bigquery/v2/rest" {
		serveDiscovery(w, r)
	} else if r.URL.Path == "/bigquery/v2/projects/speech-danstutzman/datasets" {
		serveDatasets(w, r)
	} else if r.URL.Path == "/bigquery/v2/projects/speech-danstutzman/datasets/belugacdn_logs/tables" {
		serveTables(w, r)
	} else if r.URL.Path == "/bigquery/v2/projects/speech-danstutzman/jobs" {
		startJob(w, r)
	} else if r.URL.Path == "/bigquery/v2/projects/speech-danstutzman/queries/bqjob_r7c51234c0123569f_0000015fd1968828_1" {
		serveQuery(w, r)
	} else {
		log.Fatalf("Don't know how to serve path %s", r.URL.Path)
	}
}

func main() {
	discoveryJsonPath := flag.String("discovery-json-path", "", "path to discovery.json")
	flag.Parse()

	if *discoveryJsonPath == "" {
		log.Fatalf("Please specify a -discovery-json-path")
	}

	var err error
	discoveryJson, err = ioutil.ReadFile(*discoveryJsonPath)
	if err != nil {
		panic(err)
	}
	discoveryJson = bytes.Replace(discoveryJson,
		[]byte("https://www.googleapis.com"),
		[]byte("http://localhost:9090"),
		-1)

	http.HandleFunc("/", serve)

	log.Printf("Listening on :9090...")
	err = http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
