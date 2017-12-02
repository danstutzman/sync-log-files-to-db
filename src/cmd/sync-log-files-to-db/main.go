package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/docker"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/monitis"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/redis"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3_belugacdn"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/s3_cloudtrail"
	"github.com/danielstutzman/sync-log-files-to-db/src/sources/systemd"
)

type Config struct {
	ListenOnFakeRedisForBelugaCDNLogs *redis.Options
	WatchDockerJsonFiles              *docker.Options
	WatchS3BelugaCDN                  *s3_belugacdn.Options
	WatchS3CloudTrail                 *s3_cloudtrail.Options
	WatchSystemdLogs                  *systemd.Options
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

	var wg sync.WaitGroup
	if config.ListenOnFakeRedisForBelugaCDNLogs != nil {
		redis.ValidateOptions(config.ListenOnFakeRedisForBelugaCDNLogs)

		wg.Add(1)
		go func() {
			defer wg.Done()
			redis.ListenForever(config.ListenOnFakeRedisForBelugaCDNLogs, configPath)
		}()
	}
	if config.WatchDockerJsonFiles != nil {
		docker.ValidateOptions(config.WatchDockerJsonFiles)

		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.TailDockerLogsForever(config.WatchDockerJsonFiles, configPath)
		}()
	}
	if config.WatchS3BelugaCDN != nil {
		s3_belugacdn.ValidateOptions(config.WatchS3BelugaCDN)

		wg.Add(1)
		go func() {
			defer wg.Done()
			s3_belugacdn.PollForever(config.WatchS3BelugaCDN, configPath)
		}()
	}
	if config.WatchS3CloudTrail != nil {
		s3_cloudtrail.ValidateOptions(config.WatchS3CloudTrail)

		wg.Add(1)
		go func() {
			defer wg.Done()
			s3_cloudtrail.PollForever(config.WatchS3CloudTrail, configPath)
		}()
	}
	if config.WatchSystemdLogs != nil {
		systemd.ValidateOptions(config.WatchSystemdLogs)

		wg.Add(1)
		go func() {
			defer wg.Done()
			systemd.StartTailingSystemdLogs(config.WatchSystemdLogs, configPath)
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
