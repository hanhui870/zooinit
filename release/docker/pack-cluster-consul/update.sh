# ./run.sh -d service
#   -d with debug, will not push to registry
#   -b backend
# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
#!/usr/bin/env bash

set -e

source ../Constant.rc

Cluster="cluster-consul"

docker pull ${Registry}/haimi:zooinit-${Cluster}
docker tag -f ${Registry}/haimi:zooinit-${Cluster} haimi:zooinit-${Cluster}
