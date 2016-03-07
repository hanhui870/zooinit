# ./build.sh
#   -d with debug, will not push to registry
#!/usr/bin/env bash

set -e

if [ "$1" = "-d" ]; then
    Debug="true"
else :
    Debug="false"
fi

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


# copy python scripts
SrcDir="transfer/script"
if [ -d $SrcDir ]; then
    echo "Delete tmp script dir for sync data..."
    rm -rf $SrcDir
fi
echo "Create new script dir..."
mkdir $SrcDir
#script path need under work dir
echo "Copy latest py scripts..."
cp -R ../../../script/ $SrcDir
if [ -d $BinDir"/__pycache__" ]; then
    echo "Delete bin dir __pycache__..."
    rm -rf $BinDir"/__pycache__"
fi


#compile go program
# delete bin dir
BinDir="transfer/bin"
if [ -d $BinDir ]; then
    echo "Delete tmp bin dir for sync data..."
    rm -rf $BinDir
fi
echo "Create new bin dir..."
mkdir $BinDir

imageBuild="haimi:go-docker-dev"
echo -e "Will build go program use docker container from image: "$imageBuild"..."
docker run -v /Users/bruce/:/Users/bruce/ $imageBuild bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit /Users/bruce/project/godev/src/zooinit/release/docker/pack-cluster-consul/transfer/bin"


echo -e "Will package go program into docker image...\nDir now:" `pwd`

#package code need no cache, because may change transfer files.
docker build --no-cache -t haimi:zooinit-cluster-consul .

if [ "$Debug" = "true" ]; then
    docker tag -f haimi:zooinit-cluster-consul registry.alishui.com:5000/haimi:zooinit-cluster-consul
    docker push registry.alishui.com:5000/haimi:zooinit-cluster-consul
fi
