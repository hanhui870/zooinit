#!/usr/bin/env bash

set -e

# clean dangling images
# mac not support
# docker images -q --filter "dangling=true"|awk '{print $0}' | xargs -t -i docker rmi -f {}
docker images -q --filter "dangling=true"| xargs docker rmi -f

curPwd=`pwd`

# fetch image of haimi:go-docker-dev
isExist=`docker images -q haimi:go-docker-dev | wc -l`

if [[ $isExist -lt 1 ]]; then
    cd "compiler"
    echo -e "Will build haimi:go-docker-dev images...\nDir now:" `pwd`
    docker build -t haimi:go-docker-dev .

    cd ..
fi

#compile go program
echo -e "Will build go program...\nDir now:" `pwd`
docker run -v /Users/bruce/:/Users/bruce/ haimi:go-docker-dev bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit /Users/bruce/project/godev/src/zooinit/release/docker/compiler/transfer/bin"


#package go program
cd "pack"
echo -e "Will package go program into docker image...\nDir now:" `pwd`

#package code need no cache, because may change transfer files.
docker build --no-cache -t haimi:zooinit .

cd ..

docker tag -f haimi:zooinit registry.alishui.com:5000/haimi:zooinit
docker push registry.alishui.com:5000/haimi:zooinit