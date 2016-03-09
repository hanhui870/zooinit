import subprocess
import sys
import io
import time
import json
import ipaddress
from http.client import HTTPConnection
from cluster.info import Info
from cluster.zookeeper.ServerInfo import ServerInfo, GetServerInfo
from subcall import runcmd


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    zooinfo = GetServerInfo(info)
    print("Use endpoint to detect service: " + zooinfo.GetLocalClientURL())

    while True:
        try:
            args = ["OnHealthCheck", zooinfo.GetLocalClientURL()]

            proc = runcmd.runWithStdoutSync(args)

            if proc.returncode == 0:
                print("Zookeeper Cluster is up now.")
                break
            else:
                print("Zookeeper Cluster checks up failed.")

        except Exception as err:
            print("Found error:" + str(err) + " while health check, continue loop...")

        # sleep 100ms no , 1s is enough
        time.sleep(1)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
