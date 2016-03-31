import sys
import os

import etcd

client = etcd.Client(host='registry.alishui.com', port=2379)
client.write('/nodes/n2', 2)
