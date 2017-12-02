# Tail Docker Container Logs (using `json-file` driver)

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
