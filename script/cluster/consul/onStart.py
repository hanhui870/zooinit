import subprocess
import sys
import os
from cluster.utils import printf
from cluster.info import Info


def run(info):
    if (not isinstance(info, Info)):
        printf(__name__ + "::run() info is not instance Info, please check")


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
