#构建命令：docker build --no-cache -t zooinit-cluster-zookeeper-testcluster .
#run:
#version: 0.1.0
#构建命令：docker build -t haimi:zookeeper .
#version: 0.1.0
#version: 0.1.0 基于HAIMI_ZOOKEEPER_DIST，410M。
FROM haimi:zookeeper

#MAINTAINER bruce@haimi.com

ENV PATH=$PATH:/server/zooinit/bin

ADD transfer/ /server/zooinit/

#script path need under work dir
WORKDIR /server/zooinit/

VOLUME ["/data"]

ENTRYPOINT ["/server/zooinit/entrypoint.sh"]


