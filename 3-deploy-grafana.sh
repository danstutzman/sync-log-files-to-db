#!/bin/bash -ex

rsync --rsh="ssh -i ~/.ssh/vultr" -r -z --progress \
  --no-owner backed_up/ root@build.danstutzman.com:/root/grafana/

ssh -i ~/.ssh/vultr root@build.danstutzman.com <<EOF
set -ex

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
