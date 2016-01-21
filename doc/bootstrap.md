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
