#!/usr/bin/env bash

set -e

# fetch image of haimi:go-docker-dev
isExist=`docker images -q haimi:go-docker-dev | wc -l`

if [[ $isExist -lt 1 ]]; then
    echo -e "Will build haimi:go-docker-dev images...\nDir now:" `pwd`
    docker build -t haimi:go-docker-dev .
fi
