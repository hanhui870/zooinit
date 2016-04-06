import os
from main import main

if __name__ == "__main__":
    os.environ["ZOOINIT_CLUSTER_BACKEND"] = "etcd"
    os.environ["ZOOINIT_CLUSTER_SERVICE"] = "etcd"
    os.environ["ZOOINIT_CLUSTER_EVENT"] = "OnHealthCheck"
    os.environ["ZOOINIT_SERVER_IP_LIST"] = "192.168.4.220,192.168.4.202,192.168.4.221"
    os.environ["ZOOINIT_LOCAL_IP"] = "192.168.4.220"
    os.environ["ZOOINIT_MASTER_IP"] = "192.168.4.220"
    os.environ["ZOOINIT_QURORUM"] = "3"
    os.environ[
        "ZOOINIT_SERVER_UUID_MAP"] = '{"uuu-dddd-1":"192.168.4.123","uuu-dddd-2":"192.168.4.220","uuu-dddd-3":"192.168.4.221"}'
    # os.environ["ZOOINIT_SILENT_ENV_INFO"] = "True"

    main()
