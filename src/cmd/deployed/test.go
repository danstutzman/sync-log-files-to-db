package main

import (
	"encoding/json"
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/src/storage/bigquery"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	optionsBytes, err := ioutil.ReadFile("config/bigquery.json")
	if err != nil {
		log.Fatalf("Error from ReadFile: %s", err)
	}

	var options bigquery.Options
	json.Unmarshal(optionsBytes, &options)
	bigquery.ValidateOptions(&options)

	bigqueryConn := bigquery.NewBigqueryConnection(&options)
	log.Printf("Results from SELECT 1: %v",
		bigqueryConn.Query("SELECT 1", "SELECT 1"))

	log.Printf("HELLO FROM GOLANG WITH ARGS %v\n", os.Args)
}
