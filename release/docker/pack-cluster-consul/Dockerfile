# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
#构建命令：docker build --no-cache -t haimi:zooinit-cluster-consul .
#run:
#version: 0.1.0
#version: 0.1.0 zooinit+etcd package

#need python, can't scratch
FROM haimi:base

#MAINTAINER bruce@haimi.com

ENV PATH=$PATH:/server/zooinit/bin

ADD transfer/ /server/zooinit/

#golang 安装，通过拓展PATH实现。
#20160303 need to take care of ADD order, may overwirte this.
RUN curl -o consul http://docker.alishui.com/consul && chmod a+x ./consul && mv consul /server/zooinit/bin

VOLUME ["/data"]

#script path need under work dir
WORKDIR /server/zooinit/

#EXPOSE 8300 8301 8301/udp 8302 8302/udp 8400 8500 8600 8600/udp
#Expose client port is enough. Other can also connect within docker container
EXPOSE 8500
ENTRYPOINT ["/server/zooinit/entrypoint.sh"]