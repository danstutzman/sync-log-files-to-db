package s3_belugacdn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

func readJsonIntoVisitMaps(reader io.Reader) []map[string]string {
	visits := []map[string]string{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		visit := map[string]string{}
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
