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

    url = info.GetServiceUrl(Constant.ClientPort)
    print("Use endpoint to detect service: " + url)

    while True:
        try:
            conn = HTTPConnection(url, timeout=Constant.ConnectTimeout)
            # check leader exists
            conn.request("get", "/v1/status/leader")
            resp = conn.getresponse()
            con = resp.read().decode("UTF-8").strip("")
            # json need to docode too
            leader = json.loads(con)
            print("Get leader response:" + str(resp.status) + " " + str(resp.reason) + " " + str(leader))

            conn.request("get", "/v1/status/peers")
            resp = conn.getresponse()
            peers = resp.read().decode("UTF-8").strip("")
            print("Get peers response:" + str(resp.status) + " " + str(resp.reason) + " " + str(leader))
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
                    print("Cluster status checks failed, leader is not in the peerlist.")
            else:
                print("Cluster status checks failed, leader or peers empty.")

        except Exception as err:
            print("Found error:" + str(err) + " while health check, continue loop...")

        # sleep 100ms no , 1s is enough
        time.sleep(1)

# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
