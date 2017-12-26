package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/docker"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/monitis"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/redis"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/s3_belugacdn"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/s3_cloudtrail"
	"github.com/danielstutzman/sync-log-files-to-db/src/readers/systemd"
)

type Config struct {
	LogStyle                          string
	ListenOnFakeRedisForBelugaCDNLogs *redis.Options
	TailDockerJsonFiles               *docker.Options
	PollS3BelugaCDN                   *s3_belugacdn.Options
	PollS3CloudTrail                  *s3_cloudtrail.Options
	TailSystemdLogs                   *systemd.Options
	PollMonitisResults                *monitis.Options
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

	log.Init(config.LogStyle)

	var wg sync.WaitGroup
	if config.ListenOnFakeRedisForBelugaCDNLogs != nil {
		redis.ValidateOptions(config.ListenOnFakeRedisForBelugaCDNLogs)

		wg.Add(1)
		go func() {
			defer wg.Done()
			redis.ListenForever(config.ListenOnFakeRedisForBelugaCDNLogs, configPath)
		}()
	}
	if config.TailDockerJsonFiles != nil {
		docker.ValidateOptions(config.TailDockerJsonFiles)

		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.TailDockerLogsForever(config.TailDockerJsonFiles, configPath)
		}()
	}
	if config.PollS3BelugaCDN != nil {
		s3_belugacdn.ValidateOptions(config.PollS3BelugaCDN)

		wg.Add(1)
		go func() {
			defer wg.Done()
			s3_belugacdn.PollForever(config.PollS3BelugaCDN, configPath)
		}()
	}
	if config.PollS3CloudTrail != nil {
		s3_cloudtrail.ValidateOptions(config.PollS3CloudTrail)

		wg.Add(1)
		go func() {
			defer wg.Done()
			s3_cloudtrail.PollForever(config.PollS3CloudTrail, configPath)
		}()
	}
	if config.TailSystemdLogs != nil {
		systemd.ValidateOptions(config.TailSystemdLogs)

		wg.Add(1)
		go func() {
			defer wg.Done()
			systemd.StartTailingSystemdLogs(config.TailSystemdLogs, configPath)
		}()
	}
	if config.PollMonitisResults != nil {
		monitis.ValidateOptions(config.PollMonitisResults)

		wg.Add(1)
		go func() {
			defer wg.Done()
			monitis.PollMonitisForever(config.PollMonitisResults, configPath)
		}()
	}

	wg.Wait()
}
