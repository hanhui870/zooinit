# ./run.sh -d service
#   -d with debug, will not push to registry
#   -b backend
#!/usr/bin/env bash

set -e

source ../Constant.rc

if [ "$1" = "-d" ]; then
    Debug="true"
else :
    Debug="false"
fi

if [ "$2" = "" ]; then
    service="zookeeper"
else :
    service=$2
fi

Cluster="cluster-zookeepe-testcluster"

if [ "$Debug" = "true" ]; then
    docker run -ti -P --net=host haimi:zooinit-${Cluster} zooinit cluster -b zookeeper $service

else :
    docker pull ${Registry}/haimi:zooinit-${Cluster}
    docker tag -f ${Registry}/haimi:zooinit-${Cluster} haimi:zooinit-${Cluster}

    # Use -P can expose ports to outside machine for client access.
    docker run -d -P haimi:zooinit-${Cluster} zooinit cluster -b zookeeper $service
fi
