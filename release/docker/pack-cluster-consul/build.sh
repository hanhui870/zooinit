#!/usr/bin/env bash

set -e

# clean dangling images
# mac not support
# docker images -q --filter "dangling=true"|awk '{print $0}' | xargs -t -i docker rmi -f {}
docker images -q --filter "dangling=true"| xargs docker rmi -f

# build compiler
cd "../compiler"
sh -c "./build.sh"

#package go program return dir now
Cluster="cluster-consul"
cd "../pack-"$Cluster
echo "Dir now:" `pwd`


# delete dir
BinDir="transfer/bin"
if [ -d $BinDir ]; then
    echo "Delete tmp bin dir for sync data..."
    rm -rf $BinDir
fi
echo "Create new bin dir..."
mkdir $BinDir

echo "Copy latest py scripts..."
cp -R ../../../script/ $BinDir
if [ -d $BinDir"/__pycache__" ]; then
    echo "Delete bin dir __pycache__..."
    rm -rf $BinDir"/__pycache__"
fi


#compile go program
imageBuild="haimi:go-docker-dev"
echo -e "Will build go program use docker container from image: "$imageBuild"..."
docker run -v /Users/bruce/:/Users/bruce/ $imageBuild bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit /Users/bruce/project/godev/src/zooinit/release/docker/pack-cluster-consul/transfer/bin"


echo -e "Will package go program into docker image...\nDir now:" `pwd`

#package code need no cache, because may change transfer files.
docker build --no-cache -t haimi:zooinit-cluster-consul .

cd ..

docker tag -f haimi:zooinit-cluster-consul registry.alishui.com:5000/haimi:zooinit-cluster-consul
docker push registry.alishui.com:5000/haimi:zooinit-cluster-consul