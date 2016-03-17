# Mac start

    consul cluster

# Useful command

    1. startup testcluster
        docker run -ti -P --net=host haimi:zooinit-cluster-consul zooinit cluster -b=consul -discover.target=http://192.168.4.220:2379 -ip.hint=192.168.4.108 consulTestCluster


