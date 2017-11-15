#!/bin/bash -ex

ssh -i ~/.ssh/vultr root@build.danstutzman.com <<"EOF"
set -ex

mkdir -p /etc/docker
echo '{"storage-driver":"overlay2"}' > /etc/docker/daemon.json
apt-get install -y docker.io influxdb-client

# Generate /root/influxdb.conf
if [ ! -e /root/influxdb.conf.bak ]; then
  docker run --rm influxdb influxd config > /root/influxdb.conf.bak
fi
cat /root/influxdb.conf.bak \
  | sed 's/auth-enabled = false/auth-enabled = true/g' \
  > /root/influxdb.conf

# If nothing is listening on port 8086, start InfluxDB on port 8086
mkdir -p /root/influxdb/data
lsof -i :8086 || docker run -d \
  --name=influxdb \
  -p 8086:8086 \
  -v /root/influxdb/data:/var/lib/influxdb \
  -v /root/influxdb.conf:/etc/influxdb/influxdb.conf:ro \
  --restart unless-stopped \
  influxdb -config /etc/influxdb/influxdb.conf
sleep 1

# Create InfluxDB admin user (idempotently)
INFLUXDB_PASSWORD=`cat /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/config/config.json.prod | python3 -c 'import json, sys; [print(v["InfluxDb"]["Password"]) for k, v in json.load(sys.stdin).items() if "InfluxDb" in v]' | head -1`
echo "influx -username admin -password $INFLUXDB_PASSWORD -database mydb -precision rfc3339 \"\$@\"" > /root/influx
chmod +x /root/influx
/root/influx -execute \
  "CREATE USER admin WITH PASSWORD '$INFLUXDB_PASSWORD' WITH ALL PRIVILEGES"

# If nothing is listening on port 8086, start sync-log-files-to-db on port 8086
lsof -i :6380 || docker run -d \
  --name=sync-log-files-to-db \
  -p 6380:6380 \
  -v /etc/ssl/certs:/etc/ssl/certs:ro \
  -v /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/config:/root/config:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /var/lib/docker:/var/lib/docker:ro \
  --restart unless-stopped \
  sync-log-files-to-db /root/sync-log-files-to-db /root/config/config.json.prod

mkdir -p /root/grafana/data
# If nothing is listening on port 3000, start grafana on port 3000.
lsof -i :3000 || docker run -d \
  --name=grafana \
  -p 3000:3000 \
  -v /root/grafana/data:/var/lib/grafana \
  -v /root/grafana/config:/etc/grafana \
  --restart unless-stopped \
  -e GF_INSTALL_PLUGINS=belugacdn-app \
  grafana/grafana

EOF
