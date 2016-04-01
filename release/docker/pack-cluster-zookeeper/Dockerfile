#构建命令：docker build --no-cache -t haimi:zooinit-cluster-zookeeper .
FROM haimi:zookeeper

#MAINTAINER bruce@haimi.com

ENV PATH=$PATH:/server/zooinit/bin \
    PYTHONPATH="/usr/local/python"


ADD transfer/ /server/zooinit/
# can not use absolute dir
ADD library/ ${PYTHONPATH}

RUN mkdir -p ${PYTHONPATH}

#script path need under work dir
WORKDIR /server/zooinit/

VOLUME ["/data"]

ENTRYPOINT ["/server/zooinit/entrypoint.sh"]


