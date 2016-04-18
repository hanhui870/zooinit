#构建命令：docker build --no-cache -t zooinit-cluster-etcd .
#run:
#version: 0.1.0
#version: 0.1.0 zooinit+etcd package
# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.

#need python, can't scratch
FROM haimi:base

#MAINTAINER bruce@haimi.com

ENV PATH=$PATH:/server/zooinit/bin

ADD transfer/ /server/zooinit/

#golang 安装，通过拓展PATH实现。
#20160303 need to take care of ADD order, may overwirte this.
RUN curl -o etcd-v2.2.2-linux-amd64.tar.gz http://docker.alishui.com/etcd-v2.2.2-linux-amd64.tar.gz && tar xzvf etcd-v2.2.2-linux-amd64.tar.gz && mv etcd-v2.2.2-linux-amd64/etcd* /server/zooinit/bin

VOLUME ["/data"]

#script path need under work dir
WORKDIR /server/zooinit/

EXPOSE 2379
ENTRYPOINT ["/server/zooinit/entrypoint.sh"]