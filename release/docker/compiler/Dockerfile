#构建命令：docker build -t haimi:go-docker-dev .
#run: docker run -ti -v /Users/bruce/:/Users/bruce/ haimi:go-docker-dev bash
#version: 0.1.0
#version: 0.1.0 golang+docker 开发环境,561.9 MB.
FROM haimi:go

#golang 安装，通过拓展PATH实现。
RUN mkdir -p ${HAIMI_BUILD_DIR} && cd ${HAIMI_BUILD_DIR} \
    && apt-get update &&  apt-get install -y ca-certificates curl git locate axel iptables libapparmor1 libnfnetlink0 libxtables10 --no-install-recommends \
	&& axel http://docker.alishui.com/docker-engine_1.9.1-0-trusty_amd64.deb && dpkg -i docker-engine_1.9.1-0-trusty_amd64.deb \
	&& rm -rf /var/lib/apt/lists/*  && rm -rf ${HAIMI_BUILD_DIR}

#设置环境变量 设置ENV最好都加上引号，不然网址容易出错
ENV DOCKER_TLS_VERIFY="1" \
    DOCKER_HOST="tcp://192.168.99.100:2376" \
    DOCKER_CERT_PATH="/Users/bruce/.docker/machine/machines/dev" \
    DOCKER_MACHINE_NAME="dev" \
    CGO_ENABLED=0 \
    GOPATH=$GOPATH:/Users/bruce/project/godev/

ADD transfer/ ${HAIMI_BUILD_DIR}

RUN mv ${HAIMI_BUILD_DIR}entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

CMD ["bash"]
