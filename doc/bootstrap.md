# 服务构架

1. 顶层服务基于etcd的服务发现协议.等待数量到达预订大小后,执行bootstrap.
2. 本身的顶层服务etcd也是高可用的,可以实现可以实现自举.
3. etcd发现服务启动后,可以用于启动consul, zookeeper等分布式服务.


## Bootstrap

zooinit boot|bootstrap 172.17.0.1:
