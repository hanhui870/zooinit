import subprocess
import sys
import io
import time
from cluster.info import Info
from cluster.zookeeper.ServerInfo import ServerInfo, GetServerInfo
from subcall import runcmd


# rend zookeeper config file
def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    cfgPath = "/server/zookeeper-3.4.6/conf/zoo.cfg"
    dataDir = "/data/zookeeper/data"
    idPath = dataDir + "/myid"

    zooinfo = GetServerInfo(info)

    idfile = open(idPath, "w")
    length = idfile.write(str(zooinfo.GetMyID()))
    if (length != len(str(zooinfo.GetMyID()))):
        print("Write server idfile error, length not equal to zooinfo.GetMyID()")
        sys.exit(1)

    cfgfile = open(cfgPath, "w")
    length = cfgfile.write(RenderTPL(zooinfo, dataDir))
    if (length != len(RenderTPL(zooinfo, dataDir))):
        print("Write server cfgFile error, length not equal to zooinfo.GetServerList()")
        sys.exit(1)


def RenderTPL(info, dataDir):
    if info.Qurorum != str(len(info.GetIPListArray())):
        print("info.Qurorum not equal to length of info.Iplist.")
        sys.exit(1)

    tpl = '''
# The number of milliseconds of each tick
tickTime=2000
# The number of ticks that the initial
# synchronization phase can take
initLimit=10
# The number of ticks that can pass between
# sending a request and getting an acknowledgement
syncLimit=5
# the directory where the snapshot is stored.
# do not use /tmp for storage, /tmp here is just
# example sakes.
dataDir={dataDir}
# the port at which the clients will connect
clientPort={clientPort}
# the maximum number of client connections.
# increase this if you need to handle more clients
#maxClientCnxns=60
#
# Be sure to read the maintenance section of the
# administrator guide before turning on autopurge.
#
# http://zookeeper.apache.org/doc/current/zookeeperAdmin.html#sc_maintenance
#
# The number of snapshots to retain in dataDir
autopurge.snapRetainCount=3
# Purge task interval in hours
# Set to "0" to disable auto purge feature
autopurge.purgeInterval=1

{serverlist}
'''

    return tpl.format(dataDir=dataDir, serverlist=info.GetServerList(), clientPort=info.ClientPort)

# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    # Create from info
    print("Test Create from info...")
    infoInst = Info("OnStart", "etcd", "etcd", "192.168.1.1, 192.168.1.2", "192.168.1.2", "192.168.1.2", "2")
    info = GetServerInfo(infoInst)

    if info.GetServiceUrl(8500) != "192.168.1.2:8500":
        print("Error: info.GetServiceUrl(8500)!=http://192.168.1.2:8500")

    if info.GetNodename() != "Consul-192.168.1.2":
        print("Error: info.GetNodename() != Consul-192.168.1.2")

    if info.GetMyID() != 2:
        print("info.GetMyID() found error", info.GetMyID())


    print(RenderTPL(info, "/data/zookeeper/data"))

    # run(infoInst)
