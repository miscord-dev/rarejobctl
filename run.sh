#!/bin/bash

set -eu

docker build -t rarejobctl .
docker run --platform linux/amd64 --env-file .env -it rarejobctl rarejobctl -year 2022 -month 12 -day 10 -time "10:30" -margin 30
