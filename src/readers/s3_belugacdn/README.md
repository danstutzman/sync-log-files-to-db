# Poll S3 for Beluga CDN logs

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
