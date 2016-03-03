#构建命令：docker build --no-cache -t zooinit-cluster-consul .
#run:
#version: 0.1.0
#version: 0.1.0 zooinit+etcd package

#need python, can't scratch
FROM haimi:base

MAINTAINER bruce@haimi.com

#golang 安装，通过拓展PATH实现。
RUN curl -o consul http://docker.alishui.com/consul && chmod a+x ./consul && mv consul /server/zooinit/bin

ENV PATH=$PATH:/server/zooinit/bin

ADD transfer/ /server/zooinit/

VOLUME ["/data"]

#script path need under work dir
WORKDIR /server/zooinit/

EXPOSE 8300 8301 8301/udp 8302 8302/udp 8400 8500 8600 8600/udp
ENTRYPOINT ["/server/zooinit/entrypoint.sh"]