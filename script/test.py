import os
from main import main

if __name__ == "__main__":
    os.environ["ZOOINIT_CLUSTER_BACKEND"] = "consul"
    os.environ["ZOOINIT_SERVER_IP_LIST"] = "192.168.4.108,192.168.4.114,192.168.4.144"
    os.environ["ZOOINIT_LOCAL_IP"] = "192.168.4.108"

    main()
