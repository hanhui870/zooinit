# ./build.sh
#   -d with debug, will not push to registry
#!/usr/bin/env bash

set -e

source ../Constant.rc

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
Cluster="cluster-zookeeper-test"
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

echo -e "Will build go program use docker container from image: "$ImageBuild"..."
# zooinit in the GOPATH
docker run -v ${VolumePath}:${VolumePath} $ImageBuild bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit ${ProjectPath}/release/docker/pack-${Cluster}/transfer/bin"

echo -e "Will build OnPostStart program use docker container from image: "$ImageBuild"..."
docker run -v ${VolumePath}:${VolumePath} $ImageBuild bash -c "go build -a -ldflags '-s' ${ProjectPath}/cluster/zookeeper/tools/OnPostStart.go \
    && mv OnPostStart ${ProjectPath}/release/docker/pack-${Cluster}/transfer/bin"

echo -e "Will build OnHealthCheck program use docker container from image: "$ImageBuild"..."
docker run -v ${VolumePath}:${VolumePath} $ImageBuild bash -c "go build -a -ldflags '-s' ${ProjectPath}/cluster/zookeeper/tools/OnHealthCheck.go \
    && mv OnHealthCheck ${ProjectPath}/release/docker/pack-${Cluster}/transfer/bin"


echo -e "Will package go program into docker image...\nDir now:" `pwd`

#package code need no cache, because may change transfer files.
docker build --no-cache -t haimi:zooinit-${Cluster} .

if [ "$Debug" = "false" ]; then
    docker tag -f haimi:zooinit-${Cluster} ${Registry}/haimi:zooinit-${Cluster}
    docker push ${Registry}/haimi:zooinit-${Cluster}
fi
