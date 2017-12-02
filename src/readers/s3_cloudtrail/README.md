# Poll S3 for CloudTrail logs

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
