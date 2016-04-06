# Mac start


    /Library/Java/JavaVirtualMachines/jdk1.8.0_40.jdk/Contents/Home/bin/java
        -Dzookeeper.log.dir=.
        -Dzookeeper.root.logger=INFO,CONSOLE
        -cp /Users/bruce/software/zookeeper-3.4.6/bin/../build/classes:/Users/bruce/software/zookeeper-3.4.6/bin/../build/lib/*.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/slf4j-log4j12-1.6.1.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/slf4j-api-1.6.1.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/netty-3.7.0.Final.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/log4j-1.2.16.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/jline-0.9.94.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../zookeeper-3.4.6.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../src/java/lib/*.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../conf:
        -Dcom.sun.management.jmxremote
        -Dcom.sun.management.jmxremote.local.only=false
        org.apache.zookeeper.server.quorum.QuorumPeerMain
        /Users/bruce/software/zookeeper-3.4.6/bin/../conf/zoo.cfg


# Useful command

    1. startup testcluster
        docker run -ti -P --net=host haimi:zooinit-cluster-zookeeper zooinit cluster -b=zookeeper -discovery.target=http://192.168.4.220:2379 -ip.hint=192.168.4.108 zookeeperTestCluster
        docker run -d -P --restart=always --net=host registry.alishui.com/haimi:zooinit-cluster-zookeeper zooinit cluster -b=zookeeper -discovery.target=http://192.168.4.220:2379 -ip.hint=192.168.4.108 zookeeperTestCluster

    2. normal run
          docker run -ti -P haimi:zooinit-cluster-zookeeper zooinit cluster -b=consul -discovery.target=http://192.168.4.220:2379 -ip.method=interface -ip.interface=eth0 zookeeperTmpTest1
