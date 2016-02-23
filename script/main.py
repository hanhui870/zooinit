#! /bin/env python

import sys
import os
import importlib
from cluster.utils import printf
from cluster.info import Info


# former import
# from cluster.consul import start

# this style need full path cluster.consul.start.hello()
# import cluster.consul.start

# sys.path include pwd
def main():
    printf("Zoopy started to run...")

    service = os.getenv("ZOOINIT_CLUSTER_BACKEND")
    if (service == None):
        printf("ENV ZOOINIT_CLUSTER_BACKEND is None, please check zooinit")
    else:
        printf("Receive ZOOINIT_CLUSTER_BACKEND variable: " + service)

    iplist = os.getenv("ZOOINIT_SERVER_IP_LIST")
    if (iplist == None):
        printf("ENV ZOOINIT_SERVER_IP_LIST is None, please check zooinit")
    else:
        printf("Receive ZOOINIT_SERVER_IP_LIST variable: " + iplist)

    localip = os.getenv("ZOOINIT_LOCAL_IP")
    if (localip == None):
        printf("ENV ZOOINIT_LOCAL_IP is None, please check zooinit")
    else:
        printf("Receive ZOOINIT_LOCAL_IP variable: " + localip)

    info = Info(service, iplist, localip)
    start = importlib.import_module("cluster.consul.onStart")
    start.run(info)


if __name__ == "__main__":
    main()
