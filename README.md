# Sync Log Files to Db  

## Purpose

To synchronize log data to a single database, either:
* InfluxDB (a time-series database) for fast processing
* Google BigQuery (a serverless column store) for long-term archiving and large queries

## Data sources

### Listen On Fake Redis For BelugaCDN Logs

This is a way to receive real-time logs from Beluga CDN.
They appear to arrive in less than a second in my testing.

Rather than receive BelugaCDN logs into a Redis, then scrape data out of that Redis,
`sync-log-files-to-db` understands a very limited subset of the Redis protocol,
so it acts as a "fake Redis" to recieve logs directly.

Setup:
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
