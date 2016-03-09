import sys
from cluster.info import Info
from cluster.zookeeper.ServerInfo import ServerInfo, GetServerInfo
from subcall import runcmd


# check cluster member health, un health need exit
# No need loop, Zooinit will handle it, check the exit status.
def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    zooinfo = GetServerInfo(info)
    print("Use endpoint to detect service: " + zooinfo.GetLocalClientURL())

    try:
        args = ["OnHealthCheck", zooinfo.GetLocalClientURL()]

        proc = runcmd.runWithStdoutSync(args)
        proc.wait()

        print("Zookeeper Cluster is healthy.")
        # IF error found, will trigger Exception

    except Exception as err:
        print("Found error:" + str(err) + " while health check, continue loop...")


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
