# Service bootstrap

Bootstrop the basic etcd based high available discovery service for low level use.

## Description

1. 顶层服务基于etcd的服务发现协议.等待数量到达预订大小后,执行bootstrap.
2. 本身的顶层服务etcd也是高可用的,可以实现可以实现自举.
3. etcd发现服务启动后,可以用于启动consul, zookeeper等分布式服务.


## Usage

zooinit bootstrap -f config/config.ini


## Synopsis

    1. First bootstrap etcd service of discovery in configuraion file. Then register local service self in to registry.
    2. Second boot other etcd servers in the intranet.
    3. Finally the bootstrap service is up when qurorum reach qurorum size configured in the file.

## Bootstrap

zooinit boot|bootstrap -f config/config.ini


## Bootstrap Cluster

zooinit cluster -f config/config.ini clustername

## Sample

    1. Boot initial one 192.168.4.108. If this ip want to be member of bootstrap cluster, initial discovery service need to be other ports.
        etcd -name etcd.initial.108 --data-dir /tmp/etcd/data -wal-dir /tmp/etcd/wal \
        -initial-advertise-peer-urls http://192.168.4.108:2380 \
        -listen-peer-urls http://192.168.4.108:2380 \
        -listen-client-urls http://127.0.0.1:2379,http://192.168.4.108:2379 \
        -advertise-client-urls http://192.168.4.108:2379

        //config cluster qurorum size
        curl -X PUT http://192.168.4.108:2379/v2/keys/_etcd/registry/fdsafdsafdsafdsa/_config/size -d value=${cluster_size}
    2. Boot First one 192.168.4.220
        etcd -name etcd.bootstrap.220 --data-dir /tmp/etcd/data -wal-dir /tmp/etcd/wal \
        -initial-advertise-peer-urls http://192.168.4.220:2380 \
        -listen-peer-urls http://192.168.4.220:2380 \
        -listen-client-urls http://127.0.0.1:2379,http://192.168.4.220:2379 \
        -advertise-client-urls http://192.168.4.220:2379 \
        -discovery http://192.168.4.108:2379/v2/keys/_etcd/registry/fdsafdsafdsafdsa
    2. Boot Second one 192.168.4.221
        etcd -name etcd.bootstrap.221 --data-dir /tmp/etcd/data -wal-dir /tmp/etcd/wal \
        -initial-advertise-peer-urls http://192.168.4.221:2380 \
        -listen-peer-urls http://192.168.4.221:2380 \
        -listen-client-urls http://127.0.0.1:2379,http://192.168.4.221:2379 \
        -advertise-client-urls http://192.168.4.221:2379 \
        -discovery http://192.168.4.108:2379/v2/keys/_etcd/registry/fdsafdsafdsafdsa
    3. Boot Third one 192.168.4.202
        etcd -name etcd.bootstrap.202 --data-dir /tmp/etcd/data -wal-dir /tmp/etcd/wal \
        -initial-advertise-peer-urls http://192.168.4.202:2380 \
        -listen-peer-urls http://192.168.4.202:2380 \
        -listen-client-urls http://127.0.0.1:2379,http://192.168.4.202:2379 \
        -advertise-client-urls http://192.168.4.202:2379 \
        -discovery http://192.168.4.108:2379/v2/keys/_etcd/registry/fdsafdsafdsafdsa
    4. Kill the initial one.


## QA
    Q: 2016-01-21 11:28:15.499518 E | etcdmain: member with duplicated name has registered with discovery service token(http://192.168.4.108:2379/v2/keys/_etcd/registry/fdsafdsafdsafdsa).
    A: etcd -name etcd.bootstrap.220 the name must be different.

    Q: 2016-01-21 12:13:52.381145 E | rafthttp: request sent was ignored (cluster ID mismatch: remote[e5fb7fff54887bea]=7e27652122e8b2ae, local=7021b573b4e69e)