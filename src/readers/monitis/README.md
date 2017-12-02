# Poll Monitis results

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
