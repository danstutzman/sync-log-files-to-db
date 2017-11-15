#!/bin/bash -ex

rsync --rsh="ssh -i ~/.ssh/vultr" -r -z --progress \
  --no-owner backed_up/influxdb/ root@build.danstutzman.com:/root/influxdb/

ssh -i ~/.ssh/vultr root@build.danstutzman.com <<"EOF"
set -ex

# Generate /root/influxdb.conf
mkdir -p /root/influxdb/config
if [ ! -e /root/influxdb/config/influxdb.conf.bak ]; then
  docker run --rm influxdb influxd config > /root/influxdb/config/influxdb.conf.bak
fi
cat /root/influxdb/config/influxdb.conf.bak \
  | sed 's/auth-enabled = false/auth-enabled = true/g' \
  > /root/influxdb/config/influxdb.conf

# If nothing is listening on port 8086, start InfluxDB on port 8086
mkdir -p /root/influxdb/data
lsof -i :8086 || docker run -d \
  --name=influxdb \
  -p 8086:8086 \
  -v /root/influxdb/data:/var/lib/influxdb \
  -v /root/influxdb/config:/etc/influxdb:ro \
  --restart unless-stopped \
  influxdb -config /etc/influxdb/influxdb.conf

# Create InfluxDB admin user (idempotently)
INFLUXDB_PASSWORD=`cat /root/gopath/src/github.com/danielstutzman/sync-log-files-to-db/config/config.json.prod | docker run -i --rm python:3 python -c 'import json, sys; [print(v["InfluxDb"]["Password"]) for k, v in json.load(sys.stdin).items() if "InfluxDb" in v]' | head -1`
if [ "$INFLUXDB_PASSWORD" == "" ]; then exit 1; fi
echo "docker exec \`if [ \$TERM == xterm-256color ]; then echo '-it'; fi\` influxdb influx -username admin -password $INFLUXDB_PASSWORD -database mydb -precision rfc3339 \"\$@\"" > /root/influxdb/influx
chmod +x /root/influxdb/influx
/root/influxdb/influx -execute "CREATE USER admin WITH PASSWORD '$INFLUXDB_PASSWORD' WITH ALL PRIVILEGES"

EOF
