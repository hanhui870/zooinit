import subprocess
import sys
import io
import time
import json
from http.client import HTTPConnection
from cluster.info import Info
from cluster.consul.Constant import Constant


# check cluster member health, un health need exit
def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    url = info.GetServiceUrl(Constant.ClientPort)
    print("Use endpoint to detect service: " + url)
    conn = HTTPConnection(url, timeout=Constant.ConnectTimeout)
    while True:
        # check leader exists
        conn.request("get", "/v1/health/node/" + info.GetNodename())
        resp = conn.getresponse()
        con = resp.read().decode("UTF-8").strip("")
        # json need to docode too
        health = json.loads(con)
        print("Health info " + info.GetNodename() + ":", resp.status, resp.reason, health)
        if (len(health) > 0):
            healthinfo = health[0]
            if type(healthinfo) == type({}) and "Status" in healthinfo and healthinfo["Status"] == "passing":
                print("Node " + info.GetNodename() + " health status check passing")

        time.sleep(1)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
