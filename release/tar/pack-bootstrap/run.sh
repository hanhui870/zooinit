#!/usr/bin/env bash

set -e

source ../Constant.rc

if [ "$1" = "-d" ]; then
    Debug="true"
else :
    Debug="false"
fi

if [ "$Debug" = "true" ]; then
    docker run -ti -P haimi:zooinit-bootstrap zooinit boot

else :
    docker pull ${Registry}/haimi:zooinit-bootstrap
    docker tag -f ${Registry}/haimi:zooinit-bootstrap haimi:zooinit-bootstrap

    docker run -d --net=host haimi:zooinit-bootstrap zooinit boot
fi
