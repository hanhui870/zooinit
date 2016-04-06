# ./run.sh -d service
#   -d with debug, will not push to registry
#   -b backend
#!/usr/bin/env bash

set -e

source ../Constant.rc

Cluster="cluster-zookeeper"

docker pull ${Registry}/haimi:zooinit-${Cluster}
docker tag -f ${Registry}/haimi:zooinit-${Cluster} haimi:zooinit-${Cluster}
