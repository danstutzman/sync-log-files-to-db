package s3

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
)

func (s3Connection *S3Connection) DownloadVisitsForPath(path string) []map[string]string {

	body := s3Connection.DownloadPath(path)
	reader, err := gzip.NewReader(body)
	if err != nil {
		panic(fmt.Errorf("Error from gzip.NewReader: %s", err))
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	visits := []map[string]string{}
	for scanner.Scan() {
		visit := map[string]string{
			"s3_path": path,
		}
		err := json.Unmarshal([]byte(scanner.Text()), &visit)
		if err != nil {
			panic(fmt.Errorf("Error from json.Unmarshal: %s", err))
		}

		visits = append(visits, visit)
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("Error from scanner.Err: %s", err))
	}

	return visits
}
