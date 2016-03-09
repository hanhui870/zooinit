import os
from main import main

if __name__ == "__main__":
    os.environ["ZOOINIT_CLUSTER_BACKEND"] = "zookeeper"
    os.environ["ZOOINIT_CLUSTER_SERVICE"] = "zookeeper"
    os.environ["ZOOINIT_CLUSTER_EVENT"] = "OnPostStart"
    os.environ["ZOOINIT_SERVER_IP_LIST"] = "192.168.4.108,192.168.4.220,192.168.4.221"
    os.environ["ZOOINIT_LOCAL_IP"] = "192.168.4.108"
    os.environ["ZOOINIT_MASTER_IP"] = "192.168.4.108"
    os.environ["ZOOINIT_QURORUM"] = "3"
    # os.environ["ZOOINIT_SILENT_ENV_INFO"] = "True"

    main()
