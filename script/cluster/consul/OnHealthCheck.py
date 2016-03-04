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

    while True:
        try:
            # Found error:timed out while health check, continue loop... need to create every time.
            conn = HTTPConnection(url, timeout=Constant.ConnectTimeout)
            # check leader exists
            conn.request("get", "/v1/health/node/" + info.GetNodename())
            resp = conn.getresponse()
            con = resp.read().decode("UTF-8").strip("")

            # json need to docode too
            # raise ValueError(errmsg("Expecting value", s, err.value)) from None
            if con:
                health = json.loads(con)
            else:
                health = []
            print("Health info " + info.GetNodename() + ":", resp.status, resp.reason, health)
            if (len(health) > 0):
                healthinfo = health[0]
                if type(healthinfo) == type({}) and "Status" in healthinfo and healthinfo["Status"] == "passing":
                    print("Node " + info.GetNodename() + " health status check passing")
                else:
                    print("Node health check failed.")
            else:
                print("Node health info empty.")

        except Exception as err:
            print("Found error:" + str(err) + " while health check, continue loop...")

        time.sleep(1)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
