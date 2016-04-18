# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
import sys
import traceback

from cluster.info import Info
from cluster.zookeeper.ServerInfo import ServerInfo, GetServerInfo
from subcall import runcmd


# check cluster member health, un health need exit
# No need loop, Zooinit will handle it, check the exit status.
def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    zooinfo = GetServerInfo(info)
    print("Use endpoint to detect service: " + zooinfo.GetLocalClientURL())

    try:
        args = ["OnHealthCheck", zooinfo.GetLocalClientURL()]

        proc = runcmd.runWithStdoutSync(args)
        if proc.returncode == 0:
            print("Zookeeper Cluster is healthy.")
        else:
            print("Zookeeper Cluster is NOT healthy.")
            sys.exit(1)

    except Exception as err:
        print("Found error:" + str(err) + " while health check, continue loop...")
        print(traceback.format_exc())
        sys.exit(1)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
