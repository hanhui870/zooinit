# Mac start


    /Library/Java/JavaVirtualMachines/jdk1.8.0_40.jdk/Contents/Home/bin/java
        -Dzookeeper.log.dir=.
        -Dzookeeper.root.logger=INFO,CONSOLE
        -cp /Users/bruce/software/zookeeper-3.4.6/bin/../build/classes:/Users/bruce/software/zookeeper-3.4.6/bin/../build/lib/*.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/slf4j-log4j12-1.6.1.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/slf4j-api-1.6.1.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/netty-3.7.0.Final.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/log4j-1.2.16.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../lib/jline-0.9.94.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../zookeeper-3.4.6.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../src/java/lib/*.jar:/Users/bruce/software/zookeeper-3.4.6/bin/../conf:
        -Dcom.sun.management.jmxremote
        -Dcom.sun.management.jmxremote.local.only=false
        org.apache.zookeeper.server.quorum.QuorumPeerMain
        /Users/bruce/software/zookeeper-3.4.6/bin/../conf/zoo.cfg