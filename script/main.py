#! /usr/bin/env python3
# not work #! env python

import sys
import os
import importlib
import sys
from cluster import utils
from cluster.info import Info
from cluster import signalhandler


# former import
# from cluster.consul import start

# this style need full path cluster.consul.start.hello()
# import cluster.consul.start

# sys.path include pwd
def main():
    utils.initUnbufferedStdoutIO()
    print("Zoopy started to run...")
    print("Zoopy PID: " + str(os.getpid()))
    print("PATH now: " + os.getenv("PATH"))

    print("Regist python signal handler...")
    signalhandler.registerExitSignal()

    backend = os.getenv("ZOOINIT_CLUSTER_BACKEND")
    if (backend == None):
        print("ENV ZOOINIT_CLUSTER_BACKEND is None, please check zooinit")
        sys.exit(1)
    else:
        print("Receive ZOOINIT_CLUSTER_BACKEND: " + backend)

    service = os.getenv("ZOOINIT_CLUSTER_SERVICE")
    if (service == None):
        print("ENV ZOOINIT_CLUSTER_SERVICE is None, please check zooinit")
        sys.exit(1)
    else:
        print("Receive ZOOINIT_CLUSTER_SERVICE: " + service)

    event = os.getenv("ZOOINIT_CLUSTER_EVENT")
    if (event == None):
        print("ENV ZOOINIT_CLUSTER_EVENT is None, please check zooinit")
        sys.exit(1)
    else:
        print("Receive ZOOINIT_CLUSTER_EVENT: " + event)

    iplist = os.getenv("ZOOINIT_SERVER_IP_LIST")
    if (iplist == None):
        print("ENV ZOOINIT_SERVER_IP_LIST is None, please check zooinit")
        sys.exit(1)
    else:
        print("Receive ZOOINIT_SERVER_IP_LIST: " + iplist)

    localip = os.getenv("ZOOINIT_LOCAL_IP")
    if (localip == None):
        print("ENV ZOOINIT_LOCAL_IP is None, please check zooinit")
        sys.exit(1)
    else:
        print("Receive ZOOINIT_LOCAL_IP: " + localip)

    masterip = os.getenv("ZOOINIT_MASTER_IP")
    if (masterip == None):
        print("ENV ZOOINIT_MASTER_IP is None, please check zooinit")
        sys.exit(1)
    else:
        print("Receive ZOOINIT_MASTER_IP: " + masterip)

    qurorum = os.getenv("ZOOINIT_QURORUM")
    if (qurorum == None):
        print("ENV ZOOINIT_QURORUM is None, please check zooinit")
        sys.exit(1)
    else:
        print("Receive ZOOINIT_QURORUM: " + qurorum)

    info = Info(event, service, backend, iplist, localip, masterip, qurorum)
    if (not info.CheckLocalIp()):
        print("ZOOINIT_LOCAL_IP is not in the list ZOOINIT_SERVER_IP_LIST, give up.")

    importPath = "cluster." + backend + "." + event
    print("import cluster scrpit path: " + importPath)

    try:
        start = importlib.import_module(importPath)
        start.run(info)
    except ImportError as err:
        print("Exception found:" + err.msg)
        sys.exit(1)

if __name__ == "__main__":
    main()
