#! /usr/bin/env python3
# not work #! env python

import sys
import os
import importlib
import sys
from cluster.utils import printf
from cluster.info import Info
from cluster import signalhandler


# former import
# from cluster.consul import start

# this style need full path cluster.consul.start.hello()
# import cluster.consul.start

# sys.path include pwd
def main():
    printf("Zoopy started to run...")

    printf("Regist python signal handler...")
    signalhandler.registerExitSignal()

    service = os.getenv("ZOOINIT_CLUSTER_BACKEND")
    if (service == None):
        printf("ENV ZOOINIT_CLUSTER_BACKEND is None, please check zooinit")
        sys.exit(1)
    else:
        printf("Receive ZOOINIT_CLUSTER_BACKEND variable: " + service)

    iplist = os.getenv("ZOOINIT_SERVER_IP_LIST")
    if (iplist == None):
        printf("ENV ZOOINIT_SERVER_IP_LIST is None, please check zooinit")
        sys.exit(1)
    else:
        printf("Receive ZOOINIT_SERVER_IP_LIST variable: " + iplist)

    localip = os.getenv("ZOOINIT_LOCAL_IP")
    if (localip == None):
        printf("ENV ZOOINIT_LOCAL_IP is None, please check zooinit")
        sys.exit(1)
    else:
        printf("Receive ZOOINIT_LOCAL_IP variable: " + localip)

    masterip = os.getenv("ZOOINIT_MASTER_IP")
    if (masterip == None):
        printf("ENV ZOOINIT_MASTER_IP is None, please check zooinit")
        sys.exit(1)
    else:
        printf("Receive ZOOINIT_MASTER_IP variable: " + masterip)

    info = Info(service, iplist, localip, masterip)
    if (not info.CheckLocalIp()):
        printf("ZOOINIT_LOCAL_IP is not in the list ZOOINIT_SERVER_IP_LIST, give up.")

    start = importlib.import_module("cluster.consul.onStart")
    start.run(info)


if __name__ == "__main__":
    main()
