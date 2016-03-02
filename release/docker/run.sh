#!/usr/bin/env bash

set -e

docker pull registry.alishui.com:5000/haimi:zooinit
docker tag -f registry.alishui.com:5000/haimi:zooinit haimi:zooinit

docker run -d --net=host haimi:zooinit zooinit boot