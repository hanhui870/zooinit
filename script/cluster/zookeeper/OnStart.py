import subprocess
import sys
import io
import time
from cluster.info import Info
from subcall import runcmd


# run in haimi:zookeeper container
def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    args = ["/server/zookeeper-3.4.6/bin/zkServer.sh", "start-foreground"]

    runcmd.runWithStdoutSync(args)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
