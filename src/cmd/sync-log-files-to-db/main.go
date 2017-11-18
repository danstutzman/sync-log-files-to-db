package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/docker"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/redis"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3_belugacdn"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3_cloudtrail"
)

type Config struct {
	ListenOnFakeRedisForBelugaCDNLogs *redis.Options
	WatchDockerJsonFiles              *docker.Options
	WatchS3BelugaCDN                  *s3_belugacdn.Options
	WatchS3CloudTrail                 *s3_cloudtrail.Options
}

func readConfig() (*Config, string) {
	if len(os.Args) < 1+1 {
		fmt.Fprintln(os.Stderr, "First argument should be config.json")
		os.Exit(1)
	}
	configPath := os.Args[1]

	var config = &Config{}
	configJson, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalw("Error from ioutil.ReadFile", "err", err)
	}
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		log.Fatalw("Error from json.Unmarshal", "err", err)
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
	if config.WatchS3CloudTrail != nil {
		startedOne = true
		s3_cloudtrail.ValidateOptions(config.WatchS3CloudTrail)
		go s3_cloudtrail.PollForever(config.WatchS3CloudTrail, configPath)
	}

	if startedOne {
		select {} // block forever while goroutines run
	} else {
		fmt.Fprintln(os.Stderr, "First argument should be config.json")
		os.Exit(1)
	}
}
