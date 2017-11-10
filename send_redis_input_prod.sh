#!/bin/bash -ex

cat config/redis_input.txt.prod \
  | awk '{printf "%s\r\n", $0}' \
  | netcat belugacdn-logs.danstutzman.com 6380
