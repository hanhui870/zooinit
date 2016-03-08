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
    service="consul"
else :
    service=$2
fi

if [ "$Debug" = "true" ]; then
    docker run -ti -P haimi:zooinit-cluster-consul zooinit cluster -b consul $service

else :
    docker pull registry.alishui.com:5000/haimi:zooinit-cluster-consul
    docker tag -f registry.alishui.com:5000/haimi:zooinit-cluster-consul haimi:zooinit-cluster-consul

    # Use -P can expose ports to outside machine for client access.
    docker run -d -P haimi:zooinit-cluster-consul zooinit cluster -b consul $service
fi
