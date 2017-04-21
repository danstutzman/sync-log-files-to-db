package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/src/storage/s3"
	"log"
	"strings"
)

const EXPECTED_LINE1 = "#Version: 1.0"
const EXPECTED_LINE2_V1 = "#Fields: date time x-edge-location sc-bytes c-ip cs-method cs(Host) cs-uri-stem sc-status cs(Referer) cs(User-Agent) cs-uri-query cs(Cookie) x-edge-result-type x-edge-request-id x-host-header cs-protocol cs-bytes time-taken"
const EXPECTED_LINE2_V2 = EXPECTED_LINE2_V1 + " x-forwarded-for ssl-protocol ssl-cipher x-edge-response-result-type"
const EXPECTED_LINE2_V3 = EXPECTED_LINE2_V2 + " cs-protocol-version"

func downloadVisitsForPath(s3Connection *s3.S3Connection, path string) []map[string]string {

	body := s3Connection.DownloadPath(path)
	reader, err := gzip.NewReader(body)
	if err != nil {
		panic(fmt.Errorf("Error from gzip.NewReader: %s", err))
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	if !scanner.Scan() {
		log.Fatal(fmt.Errorf("Expected at least one line of %s", path))
	}
	if scanner.Text() != EXPECTED_LINE1 {
		log.Fatal(fmt.Errorf("First line of %s should be %s but got: %s",
			path, EXPECTED_LINE1, scanner.Text()))
	}

	if !scanner.Scan() {
		log.Fatal(fmt.Errorf("Expected at least two lines of %s", path))
	}
	secondLine := scanner.Text()
	if secondLine != EXPECTED_LINE2_V1 && secondLine != EXPECTED_LINE2_V2 && secondLine != EXPECTED_LINE2_V3 {
		log.Fatal(fmt.Errorf("Expected second line of %s is: %s", path, secondLine))
	}

	visits := []map[string]string{}
	for scanner.Scan() {
		visit := map[string]string{}
		values := strings.Split(scanner.Text(), "\t")
		for i, colName := range strings.Split(secondLine, " ")[1:] {
			visit[colName] = values[i]
		}
		visits = append(visits, visit)
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("Error from scanner.Err: %s", err))
	}

	return visits
}
