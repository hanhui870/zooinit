#!/usr/bin/env bash

set -e

docker pull registry.alishui.com:5000/haimi:zooinit-cluster-consul
docker tag -f registry.alishui.com:5000/haimi:zooinit-cluster-consul haimi:zooinit-cluster-consul

# Use -P can expose ports to outside machine for client access.
docker run -d -P haimi:zooinit-cluster-consul zooinit cluster consul