import subprocess
import sys
import io
import time
import json
import ipaddress
import traceback
import etcd
from http.client import HTTPConnection
from cluster.info import Info
from cluster.etcd.Constant import Constant
from subcall import runcmd


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    url = info.GetServiceUrl(Constant.ClientPort)
    print("Use endpoint to detect service: " + url)

    while True:
        try:
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
                print("Check completed, the cluster is in sync state.")
                break
            else:
                print("Cluster status checks failed, /health info check failed.")

        except Exception as err:
            print("Found error:" + str(err) + " while health check, continue loop...")
            print(traceback.format_exc())

        # sleep 100ms no , 1s is enough
        time.sleep(1)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
