#! /usr/bin/env python3
# not work #! env python

import sys
import os
import importlib
import sys
from cluster import utils
from cluster.info import Info
from cluster import signalhandler

# add library path to autoload path
sys.path.insert(0, os.path.realpath(os.path.dirname(__file__)) + "/library")

# former import
# from cluster.consul import start

# this style need full path cluster.consul.start.hello()
# import cluster.consul.start

# sys.path include pwd
def main():
    utils.initUnbufferedStdoutIO()

    silent = os.getenv("ZOOINIT_SILENT_ENV_INFO")
    # None or nq True display debug info
    if (silent != None and silent.upper() == "TRUE"):
        silent = True
    else:
        silent = False

    if silent != True:
        print("Python sys.path:" + str(sys.path))
        print("Zoopy started to run...")
        print("Zoopy PID: " + str(os.getpid()))
        print("PATH now: " + os.getenv("PATH"))

    if silent != True:
        print("Regist python signal handler...")
    signalhandler.registerExitSignal()

    backend = os.getenv("ZOOINIT_CLUSTER_BACKEND")
    if (backend == None):
        print("ENV ZOOINIT_CLUSTER_BACKEND is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_CLUSTER_BACKEND: " + backend)

    service = os.getenv("ZOOINIT_CLUSTER_SERVICE")
    if (service == None):
        print("ENV ZOOINIT_CLUSTER_SERVICE is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_CLUSTER_SERVICE: " + service)

    event = os.getenv("ZOOINIT_CLUSTER_EVENT")
    if (event == None):
        print("ENV ZOOINIT_CLUSTER_EVENT is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_CLUSTER_EVENT: " + event)

    iplist = os.getenv("ZOOINIT_SERVER_IP_LIST")
    if (iplist == None):
        print("ENV ZOOINIT_SERVER_IP_LIST is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_SERVER_IP_LIST: " + iplist)

    uuidmap = os.getenv("ZOOINIT_SERVER_UUID_MAP")
    if (uuidmap == None):
        print("ENV ZOOINIT_SERVER_UUID_MAP is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_SERVER_UUID_MAP: " + uuidmap)


    localip = os.getenv("ZOOINIT_LOCAL_IP")
    if (localip == None):
        print("ENV ZOOINIT_LOCAL_IP is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_LOCAL_IP: " + localip)

    masterip = os.getenv("ZOOINIT_MASTER_IP")
    if (masterip == None):
        print("ENV ZOOINIT_MASTER_IP is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_MASTER_IP: " + masterip)

    qurorum = os.getenv("ZOOINIT_QURORUM")
    if (qurorum == None):
        print("ENV ZOOINIT_QURORUM is None, please check zooinit")
        sys.exit(1)
    else:
        if silent != True:
            print("Receive ZOOINIT_QURORUM: " + qurorum)

    info = Info(event, service, backend, iplist, localip, masterip, qurorum, uuidmap)
    if (not info.CheckLocalIp()):
        print("ZOOINIT_LOCAL_IP is not in the list ZOOINIT_SERVER_IP_LIST, give up.")

    importPath = "cluster." + backend + "." + event
    if silent != True:
        print("import cluster scrpit path: " + importPath)

    try:
        start = importlib.import_module(importPath)
        start.run(info)
    except ImportError as err:
        print("Exception found:" + str(err))
        sys.exit(1)


if __name__ == "__main__":
    main()
