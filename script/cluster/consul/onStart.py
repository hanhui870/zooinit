import subprocess
import sys
import os

from cluster.utils import printf
from cluster.info import Info


def run(info):
    if (not isinstance(info, Info)):
        printf(__name__ + "::run() info is not instance Info, please check")

    try:
        # fail, no quoted: consul agent -server -data-dir="/tmp/consul" -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108
        output = subprocess.check_output([
                                             "consul agent -server -data-dir=/tmp/consul -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108"],
                                         stderr=subprocess.STDOUT, shell=True)
        printf(output)
    except Exception as err:
        printf("Found error:", err)

# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
