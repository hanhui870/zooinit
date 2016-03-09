# ./run.sh -d service
#   -d with debug, will not push to registry
#   -b backend
#!/usr/bin/env bash

set -e

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

Cluster="cluster-zookeeper"

if [ "$Debug" = "true" ]; then
    docker run -ti -P haimi:zooinit-${Cluster} zooinit cluster -b zookeeper $service

else :
    docker pull registry.alishui.com:5000/haimi:zooinit-${Cluster}
    docker tag -f registry.alishui.com:5000/haimi:zooinit-${Cluster} haimi:zooinit-${Cluster}

    # Use -P can expose ports to outside machine for client access.
    docker run -d -P haimi:zooinit-${Cluster} zooinit cluster -b zookeeper $service
fi
