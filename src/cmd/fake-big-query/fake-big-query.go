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
	fmt.Println("serveDiscovery", r.URL.Path)
	w.Write(discoveryJson)
}

func serveDatasets(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serveDatasets", r.URL.Path)
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

	http.HandleFunc("/discovery/v1/apis/bigquery/v2/rest", serveDiscovery)
	http.HandleFunc("/bigquery/v2/projects/speech-danstutzman/datasets", serveDatasets)

	log.Printf("Listening on :9090...")
	err = http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
