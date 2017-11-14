#!/bin/bash -ex

INFLUXDB_PASSWORD=`cat config/INFLUXDB_PASSWORD`

ssh -i ~/.ssh/vultr root@build.danstutzman.com <<EOF
set -ex

docker ps -a -f ancestor=google/cadvisor:latest --format={{.ID}} \
    | xargs --no-run-if-empty docker stop
docker ps -a -f ancestor=google/cadvisor:latest --format={{.ID}} \
    | xargs --no-run-if-empty docker rm

/root/influx -execute "CREATE DATABASE cadvisor"
docker run -d \
  --name=cadvisor \
  -p 8080:8080 \
  -v /var/run:/var/run:rw \
  -v /sys:/sys:ro \
  -v /var/lib/docker/:/var/lib/docker:ro \
  -v /var/log:/var/log:ro \
  --restart unless-stopped \
  google/cadvisor:latest \
    -storage_driver=influxdb \
    -storage_driver_host=build.danstutzman.com:8086 \
    -storage_driver_user=admin \
    -storage_driver_password=$INFLUXDB_PASSWORD \
    -docker_only=true \
    -logtostderr=true

EOF
