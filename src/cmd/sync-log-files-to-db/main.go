package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/danielstutzman/sync-log-files-to-db/src/sources/docker"
	my_s3 "github.com/danielstutzman/sync-log-files-to-db/src/sources/s3"
)

type Config struct {
	WatchDockerJsonFiles *docker.Options
	WatchS3              *my_s3.Options
}

func readConfig() (*Config, string) {
	if len(os.Args) < 1+1 {
		log.Fatalf("First argument should be config.json")
	}
	configPath := os.Args[1]

	var config = &Config{}
	configJson, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error from ioutil.ReadFile: %s", err)
	}
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		log.Fatalf("Error from json.Unmarshal: %s", err)
	}
	return config, configPath
}

func main() {
	config, configPath := readConfig()

	if config.WatchDockerJsonFiles == nil && config.WatchS3 == nil {
		log.Fatalf("Config should include one or both of WatchDockerJsonFiles and WatchS3")
	}

	if config.WatchDockerJsonFiles != nil {
		docker.ValidateOptions(config.WatchDockerJsonFiles)
		go docker.TailDockerLogsForever(config.WatchDockerJsonFiles, configPath)
	}

	if config.WatchS3 != nil {
		my_s3.ValidateOptions(config.WatchS3)
		go my_s3.PollForever(config.WatchS3, configPath)
	}

	select {} // block forever while goroutines run
}
