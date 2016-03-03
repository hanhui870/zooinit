#!/usr/bin/env bash

set -e

# clean dangling images
# mac not support
# docker images -q --filter "dangling=true"|awk '{print $0}' | xargs -t -i docker rmi -f {}
docker images -q --filter "dangling=true"| xargs docker rmi -f

# build compiler
cd "../compiler"
sh -c "./build.sh"


#compile go program
imageBuild="haimi:go-docker-dev"
echo -e "Will build go program use docker container from image: "$imageBuild"..."
docker run -v /Users/bruce/:/Users/bruce/ $imageBuild bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit /Users/bruce/project/godev/src/zooinit/release/docker/pack-cluster-consul/transfer/bin"


#package go program return dir now
cd "../pack-cluster-consul"
echo -e "Will package go program into docker image...\nDir now:" `pwd`

#package code need no cache, because may change transfer files.
docker build --no-cache -t haimi:zooinit-cluster-consul .

cd ..

docker tag -f haimi:zooinit-cluster-consul registry.alishui.com:5000/haimi:zooinit-cluster-consul
docker push registry.alishui.com:5000/haimi:zooinit-cluster-consul