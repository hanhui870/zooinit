# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
import subprocess
import sys
import io
import time
from cluster.info import Info
from subcall import runcmd
from cluster.etcd.Constant import Constant


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")
        sys.exit(1)

    # see https://github.com/coreos/etcd/blob/master/Documentation/clustering.md
    args = ["etcd"]

    args.append("--data-dir=/tmp/etcd11/data")
    args.append("--wal-dir=/tmp/etcd11/wal")
    args.append("--name=" + info.GetNodename())
    args.append("--initial-advertise-peer-urls=http://" + info.GetServiceUrl(Constant.PeerPort))
    args.append("--listen-peer-urls=http://" + info.GetServiceUrl(Constant.PeerPort))
    args.append("--advertise-client-urls=http://" + info.GetServiceUrl(Constant.ClientPort))
    args.append("--listen-client-urls=http://127.0.0.1:2379,http://" + info.GetServiceUrl(Constant.ClientPort))
    args.append("--initial-cluster-token=" + info.Service)

    ips = info.GetIPListArray()
    purls = list()
    for ip in ips:
        purls.append(info.GetNodenameOfNode(ip) + "=http://" + ip + ":" + str(Constant.PeerPort))

    args.append("--initial-cluster=" + ",".join(purls))
    args.append("--initial-cluster-state=new")


    runcmd.runWithStdoutSync(args)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
