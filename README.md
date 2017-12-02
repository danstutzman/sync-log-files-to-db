# Sync Structured Log Files to a Time-series Database

## Purpose

To synchronize log data to a single database, either:
* InfluxDB (a time-series database) for fast processing
* Google BigQuery (a serverless column store) for long-term archiving and large queries

## Data sources

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
      "DatabaseName": "(choose an InfluxDB database name)",
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
