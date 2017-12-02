# Sync Structured Log Files to a Time-series Database

## Purpose

To synchronize log data to a single database, either:
* InfluxDB (a time-series database) for fast processing
* Google BigQuery (a serverless column store) for long-term archiving and large queries

## Data sources

### Tail Docker Container Logs (using `json-file` driver)

Docker saves its container's output using the logging driver you specify.  The default option is [JSON file logging](https://docs.docker.com/engine/admin/logging/json-file/).  Each line of stdout or stderr from your docker container is appended as a JSON object including a timestamp.

How to set it up:
- `docker run`'s `--log-driver` option should default to `json-file`, but maybe check to make sure you're not passing ` -–log-driver syslog`
- Add the following stanza to your config.json:
  ```
  "TailDockerJsonFiles": {
    "SecondsBetweenPollForNewContainers": 60,
    "InfluxDb": {
      "Hostname": "127.0.0.1",
      "Port": "8086",
      "DatabaseName": "mydb",
      "MeasurementName": "docker_logs"
    }
  },
  ```

If a log line appears to match Apache combined-log output, it will be parsed into separate fields.

### Poll Monitis results

Monitis is a third-party service to regularly send HTTP requests to your web site to see if it's still up.
Besides alerting you if your site is down, they'll log how long your web site took to respond.
`sync-log-files-to-db` lets you collect these logs without having to manually login to the Monitis dashboard and export CSV files.

How to set it up:
- Login to Monitis dashboard
- Create external monitors if you haven't already
- Go to Tools -> API -> API Key to look up your API key and Secret key
- Add the following stanza to your config.json:
  ```
  "PollMonitisResults": {
    "ApiKey": "YOUR-MONITIS-API-KEY-HERE",
    "SecretKey": "YOUR-MONITIS-SECRET-KEY-HERE",
    "InfluxDb": {
      "Hostname": "127.0.0.1",
      "Port": "8086",
      "DatabaseName": "mydb",
      "MeasurementName": "monitis_results"
    }
  }
  ```

### Poll S3 for CloudTrail logs

If you enable AWS CloudTrail, AWS will log every AWS API call to a S3 bucket of your choice.

How to set it up:
- Go to AWS dashboard for CloudTrail, and create a "trail" with the following options:
  - Encrypt log files = No (since `sync-log-files-to-db` doesn't know how to decrypt them)
  - Enable log file validation = No (since `sync-log-files-to-db` won't bother to delete the checksum files, so they'll just accumulate)
- Create an AWS S3 bucket for your CloudTrail logs
- Create an AWS IAM user that has read and delete access to just that S3 bucket
- Setup a `config/s3.creds.ini` file using `config/s3.creds.ini.sample` as a starting point
- Add the following stanza to your config.json (omitting BigQuery or InfluxDb if you prefer):
  ```
  "PollS3CloudTrail": {
    "S3": {
      "CredsPath": "./s3.creds.ini",
      "Region": "us-east-1",
      "BucketName": "cloudtrail-danstutzman"
    },
    "PathsPerBatch": 100,

    "BigQuery": {
      "GcloudPemPath": "./YourProject-abc123.json",
      "GcloudProjectId": "your-project",
      "DatasetName": "cloudtrail",
      "TableName": "cloudtrail_events"
    },
    "InfluxDb": {
      "Hostname": "127.0.0.1",
      "Port": "8086",
      "DatabaseName": "mydb",
      "MeasurementName": "cloudtrail_events"
    }
  },
  ```

### Poll S3 for Beluga CDN logs

BelugaCDN can [send their logs to an S3 bucket of your choice](https://docs.belugacdn.com/docs/real-time).
They appear to arrive every five minutes in my testing, whether your BelugaCDN site had any accesses or not.

How to set it up:
- Create an S3 bucket
- Create an AWS IAM user that has only `s3:PutObject` access to `arn:aws:s3:::YOUR_BUCKET/*`
- Put the access key and secret access key in a file like `config/s3.creds.ini`,
  using `config/s3.creds.ini.sample` as a starting point
- Add the following stanza to your config.json (omitting BigQuery or InfluxDb if you prefer):
  ```
  "PollS3BelugaCDN": {
    "S3": {
      "CredsPath": "./s3.creds.ini",
      "Region": "us-east-1",
      "BucketName": "belugacdn-logs-danstutzman"
    },
    "PathsPerBatch": 100,

    "BigQuery": {
      "GcloudPemPath": "./YourProjectName-123abc.json",
      "GcloudProjectId": "your-project-id",
      "DatasetName": "belugacdn_logs",
      "TableName": "visits"
    },
    "InfluxDb": {
      "Hostname": "127.0.0.1",
      "Port": "8086",
      "DatabaseName": "mydb",
      "MeasurementName": "belugacdn_logs"
    }
  }
  ```
- File a ticket with BelugaCDN asking to enable S3 logs, including the following info:
  * **bucket name**
  * **aws_region** for the S3 bucket
  * **path** (or "/" if you don't care)
  * **user_id** for the IAM user you created
  * **access_key** for the IAM user you created
- Wait for BelugaCDN to respond

### Listen On Fake Redis For BelugaCDN Logs

This is a way to receive [real-time logs from Beluga CDN](https://docs.belugacdn.com/docs/real-time).
They appear to arrive in less than a second in my testing.

Rather than receive BelugaCDN logs into a Redis, then scrape data out of that Redis,
`sync-log-files-to-db` understands a very limited subset of the Redis protocol,
so it acts as a "fake Redis" to recieve logs directly.

How to set it up:
- Add the following stanza to your config.json:
   ```
  "ListenOnFakeRedisForBelugaCDNLogs": {
    "ListenPort": "6380",
    "ExpectedPassword": "(choose a password and put it here)",
    "InfluxDb": {
      "Hostname": "127.0.0.1",
      "Port": "8086",
      "DatabaseName": "mydb",
      "MeasurementName": "belugacdn_logs"
    }
  }
  ```
- Test with `send_redis_input.sh`
- Open your firewall so the fake-Redis port (e.g. 6380) is accessible from the outside
- Test with `send_redis_input_prod.sh`
- File a ticket with BelugaCDN asking to enable Redis logs, including the following info:
  * **hostname** for your running sync-log-files-to-db daemon
  * **port** for fake Redis (`ListenPort`)
  * **password** for fake Redis (`ExpectedPassword`)
- Wait for BelugaCDN to respond
