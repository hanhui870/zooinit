import subprocess
import sys
import io
import time
from cluster.info import Info
from subcall import runcmd


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

        # print("Hello world: "+info.Event)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
