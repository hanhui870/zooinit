# Mac start

    consul cluster

# Useful command

    1. startup testcluster
        docker run -ti -P --net=host haimi:zooinit-cluster-consul zooinit cluster -b=consul -discovery.target=http://192.168.4.220:2379 -ip.hint=192.168.4.108 consulTestCluster
        docker run -d -P --restart=always --net=host registry.alishui.com/haimi:zooinit-cluster-consul zooinit cluster -b=consul -discovery.target=http://192.168.4.220:2379 -ip.hint=192.168.4.108 consulTestCluster
    2. normal test
        docker run -ti -P haimi:zooinit-cluster-consul zooinit cluster -b=consul -discovery.target=http://192.168.4.220:2379 consulTmpTest1

