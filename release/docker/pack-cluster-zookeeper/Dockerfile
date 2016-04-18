#构建命令：docker build --no-cache -t haimi:zooinit-cluster-zookeeper .
# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
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


