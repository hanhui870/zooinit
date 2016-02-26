import subprocess
import sys
import io
from cluster.info import Info
from subcall import runcmd


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    args = ["consul", "agent",
            "-node=Consul-" + info.Localip,
            "-data-dir=/tmp/consul",
            "-bind=" + info.Localip,
            "-client=" + info.Localip]

    if (info.Localip != info.Masterip):  # slave mode
        args.append("-join=" + info.Masterip)

    else:  # server mode
        args.append("-server")

        args.append("-bootstrap-expect")
        args.append(info.Qurorum)

    runcmd.runWithStdoutSync(args)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
