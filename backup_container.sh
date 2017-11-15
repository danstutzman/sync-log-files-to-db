#!/bin/bash -ex

mkdir -p backed_up
rsync -v -r --relative \
  root@build.danstutzman.com:/root/grafana/ \
  root@build.danstutzman.com:/root/influxdb/ \
  root@build.danstutzman.com:/root/last_influxdb_log.sh \
  root@build.danstutzman.com:/root/last_belugacdn_log.sh \
  backed_up
