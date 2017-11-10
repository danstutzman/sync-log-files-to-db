#!/bin/bash -ex

ssh -i ~/.ssh/vultr root@build.danstutzman.com <<"EOF"
set -ex

apt-get install -y docker.io influxdb-client

# Generate /root/influxdb.conf
if [ ! -e /root/influxdb.conf.bak ]; then
  docker run --rm influxdb influxd config > /root/influxdb.conf.bak
fi
cat /root/influxdb.conf.bak \
  | sed 's/auth-enabled = false/auth-enabled = true/g' \
  > /root/influxdb.conf

# If nothing is listening on port 8086, start InfluxDB on port 8086
lsof -i :8086 || docker run -d \
  -p 8086:8086 \
  -v /root/influxdb.conf:/etc/influxdb/influxdb.conf:ro \
  --restart unless-stopped \
  influxdb -config /etc/influxdb/influxdb.conf && sleep 1

# Create InfluxDB admin user (idempotently)
INFLUXDB_PASSWORD=`cat /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/config/config.json.prod | python3 -c 'import json, sys; [print(v["InfluxDb"]["Password"]) for k, v in json.load(sys.stdin).items() if "InfluxDb" in v]' | head -1`
echo "influx -username admin -password $INFLUXDB_PASSWORD \"\$@\"" > /root/influx
chmod +x /root/influx
/root/influx -execute \
  "CREATE USER admin WITH PASSWORD '$INFLUXDB_PASSWORD' WITH ALL PRIVILEGES"

# Build sync-log-files-to-db Docker image
cp /root/gopath/bin/sync-log-files-to-db \
  /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/sync-log-files-to-db
docker build /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db \
  -t sync-log-files-to-db
rm -f /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/sync-log-files-to-db

# If nothing is listening on port 8086, start sync-log-files-to-db on port 8086
lsof -i :6380 || docker run -d \
  -p 6380:6380 \
  -v /etc/ssl/certs:/etc/ssl/certs:ro \
  -v /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/config:/root/config:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --restart unless-stopped \
  sync-log-files-to-db /root/sync-log-files-to-db /root/config/config.json.prod

EOF
