#!/usr/bin/env bash

set -e

source ../Constant.rc

# clean dangling images
# mac not support
# docker images -q --filter "dangling=true"|awk '{print $0}' | xargs -t -i docker rmi -f {}
docker images -q --filter "dangling=true"| xargs docker rmi -f

# build compiler
cd "../compiler"
sh -c "./build.sh"

#package go program return dir now
Cluster="bootstrap"
cd "../pack-"$Cluster
echo "Dir now:" `pwd`


#compile go program
echo -e "Will build go program use docker container from image: "$ImageBuild"..."
docker run -v ${VolumePath}:${VolumePath} $ImageBuild bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit ${ProjectPath}/release/docker/pack-${Cluster}/transfer/bin"


#package go program return dir now
echo -e "Will package go program into docker image...\nDir now:" `pwd`

#package code need no cache, because may change transfer files.
docker build --no-cache -t haimi:zooinit-${Cluster} .

cd ..

docker tag -f haimi:zooinit-${Cluster} ${Registry}/haimi:zooinit-${Cluster}
docker push ${Registry}/haimi:zooinit-${Cluster}