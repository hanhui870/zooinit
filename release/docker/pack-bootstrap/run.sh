#!/usr/bin/env bash

set -e

docker pull registry.alishui.com:5000/haimi:zooinit-bootstrap
docker tag -f registry.alishui.com:5000/haimi:zooinit-bootstrap haimi:zooinit-bootstrap

docker run -d --net=host haimi:zooinit-bootstrap zooinit boot