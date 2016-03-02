import subprocess
import sys
import io
import time
import json
import ipaddress
from http.client import HTTPConnection
from cluster.info import Info
from cluster.consul.Constant import Constant
from subcall import runcmd


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    global ClientPort, ConnectTimeout
    url = info.GetServiceUrl(Constant.ClientPort)
    print("Use endpoint to detect service: " + url)
    conn = HTTPConnection(url, timeout=Constant.ConnectTimeout)
    while True:
        # check leader exists
        conn.request("get", "/v1/status/leader")
        resp = conn.getresponse()
        con = resp.read().decode("UTF-8").strip("")
        # json need to docode too
        leader = json.loads(con)
        print("Get response:", resp.status, resp.reason, leader)

        conn.request("get", "/v1/status/peers")
        resp = conn.getresponse()
        peers = resp.read().decode("UTF-8").strip("")
        print("Get response:", resp.status, resp.reason, peers)
        if leader != "" and peers != "":
            peerlist = json.loads(peers)
            if leader in peerlist:
                # check leader format
                leaderip = leader[:str.find(leader, ":")]
                try:
                    ipv4 = ipaddress.IPv4Address(leaderip)
                    print("Leader ip check pass: " + str(leaderip))

                    print("Check completed, the cluster is in sync state.")
                    break
                except Exception as err:
                    print("Found error:" + str(err) + " will quit health check.")
                    sys.exit(1)

            else:
                pass

        # sleep 100ms
        time.sleep(1 / 10)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
