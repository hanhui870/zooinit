#!/usr/bin/env bash

set -e

source ././../../docker/Constant.rc

# clean dangling images
# mac not support
# docker images -q --filter "dangling=true"|awk '{print $0}' | xargs -t -i docker rmi -f {}
docker images -q --filter "dangling=true"| xargs docker rmi -f

# build compiler
cd "././../../docker/compiler"
sh -c "./build.sh"

#package go program return dir now
Cluster="bootstrap"
cd -
echo "Dir now:" `pwd`


#compile go program
echo -e "Will build go program use docker container from image: "$ImageBuild"..."
docker run -v ${VolumePath}:${VolumePath} $ImageBuild bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit ${ProjectPath}/release/tar/pack-${Cluster}/transfer/bin"

# Download etcd file
axel http://docker.alishui.com/etcd-v2.2.2-linux-amd64.tar.gz && tar xzvf etcd-v2.2.2-linux-amd64.tar.gz \
    && mv etcd-v2.2.2-linux-amd64/etcd* ${ProjectPath}/release/tar/pack-${Cluster}/transfer/bin && rm -rf etcd-v2.2.2*

# package go program return dir now
echo -e "Will package go program into tar file...\nDir now:" `pwd`

cp -a transfer/ outupt/zooinit-${Version}
tar -czf outupt/zooinit-${Version}.tar.gz zooinit-${Version}/

#TODO upload file to hosts.



