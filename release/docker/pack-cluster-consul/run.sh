#!/usr/bin/env bash

set -e

docker pull registry.alishui.com:5000/haimi:zooinit-cluster-consul
docker tag -f registry.alishui.com:5000/haimi:zooinit-cluster-consul haimi:zooinit-cluster-consul

docker run -d --net=host haimi:zooinit-cluster-consul zooinit cluster consul