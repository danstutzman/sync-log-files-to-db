package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/danielstutzman/sync-log-files-to-db/src/sources/docker"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/redis"
	my_s3 "github.com/danielstutzman/sync-log-files-to-db/src/sources/s3"
)

type Config struct {
	ListenOnFakeRedisForBelugaCDNLogs *redis.Options
	WatchDockerJsonFiles              *docker.Options
	WatchS3                           *my_s3.Options
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

	startedOne := false
	if config.ListenOnFakeRedisForBelugaCDNLogs != nil {
		startedOne = true
		redis.ValidateOptions(config.ListenOnFakeRedisForBelugaCDNLogs)
		go redis.ListenForever(config.ListenOnFakeRedisForBelugaCDNLogs, configPath)
	}
	if config.WatchDockerJsonFiles != nil {
		startedOne = true
		docker.ValidateOptions(config.WatchDockerJsonFiles)
		go docker.TailDockerLogsForever(config.WatchDockerJsonFiles, configPath)
	}
	if config.WatchS3 != nil {
		startedOne = true
		my_s3.ValidateOptions(config.WatchS3)
		go my_s3.PollForever(config.WatchS3, configPath)
	}

	if startedOne {
		select {} // block forever while goroutines run
	} else {
		log.Fatalf("No sources specified in config")
	}
}
