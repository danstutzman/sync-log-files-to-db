package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/danielstutzman/sync-log-files-to-db/src/sources/docker"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/redis"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3_belugacdn"
)

type Config struct {
	ListenOnFakeRedisForBelugaCDNLogs *redis.Options
	WatchDockerJsonFiles              *docker.Options
	WatchS3BelugaCDN                  *s3_belugacdn.Options
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
	if config.WatchS3BelugaCDN != nil {
		startedOne = true
		s3_belugacdn.ValidateOptions(config.WatchS3BelugaCDN)
		go s3_belugacdn.PollForever(config.WatchS3BelugaCDN, configPath)
	}

	if startedOne {
		select {} // block forever while goroutines run
	} else {
		log.Fatalf("No sources specified in config")
	}
}
