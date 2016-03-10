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
    service="consul"
else :
    service=$2
fi

Cluster="cluster-consul"

if [ "$Debug" = "true" ]; then
    docker run -ti -P haimi:zooinit-${Cluster} zooinit cluster -b consul $service

else :
    docker pull ${Registry}/haimi:zooinit-${Cluster}
    docker tag -f ${Registry}/haimi:zooinit-${Cluster} haimi:zooinit-${Cluster}

    # Use -P can expose ports to outside machine for client access.
    docker run -d -P haimi:zooinit-${Cluster} zooinit cluster -b consul $service
fi
