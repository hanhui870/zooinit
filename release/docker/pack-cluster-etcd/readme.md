# Mac start

    etcd cluster

# Useful command

    1. startup testcluster
        docker run -ti -P --net=host haimi:zooinit-cluster-etcd zooinit cluster -b=etcd -discovery.target=http://192.168.4.220:2379 -ip.hint=192.168.4.108 etcdTestCluster
        docker run -d -P --restart=always --net=host registry.alishui.com/haimi:zooinit-cluster-etcd zooinit cluster -b=etcd -discovery.target=http://192.168.4.220:2379 -ip.hint=192.168.4.108 etcdTestCluster
    2. normal test
        docker run -ti -P haimi:zooinit-cluster-etcd zooinit cluster -b=etcd -discovery.target=http://192.168.4.220:2379 -ip.method=interface -ip.interface=eth0 etcdTmpTest1

