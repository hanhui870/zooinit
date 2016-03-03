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
echo -e "Will build go program...\nDir now:" `pwd`
docker run -v /Users/bruce/:/Users/bruce/ haimi:go-docker-dev bash -c "go build -a -ldflags '-s' zooinit \
    && mv zooinit /Users/bruce/project/godev/src/zooinit/release/docker/pack-bootstrap/transfer/bin"


#package go program return dir now
cd -
echo -e "Will package go program into docker image...\nDir now:" `pwd`

#package code need no cache, because may change transfer files.
docker build --no-cache -t haimi:zooinit-bootstrap .

cd ..

docker tag -f haimi:zooinit-bootstrap registry.alishui.com:5000/haimi:zooinit-bootstrap
docker push registry.alishui.com:5000/haimi:zooinit-bootstrap