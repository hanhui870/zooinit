# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
import subprocess
import sys
import io
import time
import json
import traceback
from http.client import HTTPConnection
from cluster.info import Info
from cluster.etcd.Constant import Constant


# check cluster member health, un health need exit
# No need loop, Zooinit will handle it, check the exit status.
def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    url = info.GetServiceUrl(Constant.ClientPort)
    print("Use endpoint to detect service: " + url)

    try:
        # Found error:timed out while health check, continue loop... need to create every time.
        conn = HTTPConnection(url, timeout=Constant.ConnectTimeout)
        # check health info, etcd case sensitive
        conn.request("GET", "/health")
        resp = conn.getresponse()
        con = resp.read().decode("UTF-8").strip("")

        # json need to docode too
        try:
            health = json.loads(con)
        except Exception as err:
            health = ""

        # str(resp.headers)
        print("Get health response:" + str(resp.status) + " " + str(resp.reason) + " " + str(health))

        if (con != "" and isinstance(health, object) and health["health"] == "true"):
            print("Node " + info.GetNodename() + " health status check passing")
        else:
            print("Cluster health checks failed, /health info check failed.")
            sys.exit(1)

    except Exception as err:
        print("Found error:" + str(err) + " while health check, continue loop...")
        print(traceback.format_exc())
        sys.exit(1)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
