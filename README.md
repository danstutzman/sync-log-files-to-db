[![GoDoc](https://godoc.org/github.com/danielstutzman/sync-log-files-to-db?status.svg)](https://godoc.org/github.com/danielstutzman/sync-log-files-to-db/src/cmd/sync-log-files-to-db?imports)

# Sync Structured Log Files to a Time-series Database

## Purpose

To consolidate incoming logs into a single time-series database

## Readers (where `sync-log-files-to-db` can read logs from)
* [Docker container logs](src/readers/docker/README.md)
* [Monitis results](src/readers/monitis/README.md)
* [BelugaCDN real-time logs](src/readers/redis/README.md)
* [BelugaCDN logs in S3](src/readers/s3_belugacdn/README.md)
* [AWS CloudTrail logs in S3](src/readers/s3_cloudtrail/README.md)
* Systemd journal logs

## Writers (where `sync-log-files-to-db` can write logs to)
* InfluxDB, a time-series database
* Google BigQuery, a serverless "big data" column store, useful for long-term archiving and large queries
