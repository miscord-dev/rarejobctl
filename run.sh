#!/bin/bash

set -eu

docker build -t rarejobctl-standalone --file Dockerfile-standalone .
docker run --platform linux/amd64 --env-file .env -it rarejobctl-standalone rarejobctl -year 2022 -month 12 -day 27 -time "9:30" -margin 30 -debug true
