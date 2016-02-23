import subprocess
import sys
import os


def run():
    print("Zoopy started to run...")
    iplist = os.getenv("ZOOINIT_SERVER_IP_LIST")
    if (iplist == None):
        print("ENV ZOOINIT_SERVER_IP_LIST is None, please check zooinit")
