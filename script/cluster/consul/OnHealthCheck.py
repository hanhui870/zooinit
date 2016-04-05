import subprocess
import sys
import io
import time
import json
import traceback
from http.client import HTTPConnection
from cluster.info import Info
from cluster.consul.Constant import Constant


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
        # check leader exists
        conn.request("get", "/v1/health/node/" + info.GetBackendNodename())
        resp = conn.getresponse()
        con = resp.read().decode("UTF-8").strip("")

        # json need to docode too
        # raise ValueError(errmsg("Expecting value", s, err.value)) from None
        try:
            health = json.loads(con)
        except Exception as err:
            health = []
        print("Health info " + info.GetBackendNodename() + ":" + str(resp.status) + " " + str(resp.reason) + " " + str(
            health))
        if (len(health) > 0):
            healthinfo = health[0]
            if type(healthinfo) == type({}) and "Status" in healthinfo and healthinfo["Status"] == "passing":
                print("Node " + info.GetBackendNodename() + " health status check passing")
            else:
                print("Node health check failed.")
                sys.exit(1)
        else:
            print("Node health info empty.")
            sys.exit(1)

    except Exception as err:
        print("Found error:" + str(err) + " while health check, continue loop...")
        print(traceback.format_exc())
        sys.exit(1)



# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
