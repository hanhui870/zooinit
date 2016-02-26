import subprocess
import sys
import io
from cluster.info import Info
from subcall import runcmd


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    runcmd.runWithStdoutSync(["consul", "agent", "-server",
                              "-data-dir=/tmp/consul",
                              "-bootstrap-expect", info.Qurorum,
                              "-bind=" + info.Localip,
                              "-client=" + info.Localip])


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
