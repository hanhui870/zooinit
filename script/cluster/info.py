class Info(object):
    'Cluster Boot info'

    def __init__(self, event, service, backend, iplist, localip, masterip, qurorum, uuidmap=None):
        self.Event = event
        self.Service = service
        self.Backend = backend
        self.Iplist = iplist
        self.Localip = localip
        self.Masterip = masterip
        self.Qurorum = qurorum
        # this is string
        self.UUIDMap = uuidmap

    def GetIPListArray(self):
        list = self.Iplist.split(",")

        result = []
        for unit in list:
            result.append(unit.strip(" \t"))

        return result

    def CheckLocalIp(self):
        list = self.GetIPListArray()
        for ip in list:
            if ip == self.Localip:
                return True
        return False

    def GetServiceUrl(self, port):
        # python http.client protocal
        return self.Localip + ":" + str(port)

    def GetNodename(self):
        return self.Service + "-" + self.Localip

    # backend nodename, such as consul need its backend name.
    def GetBackendNodename(self):
        return self.Backend.capitalize() + "-" + self.Localip

    def GetUUIDMap(self):
        # Json to map
        list = self.Iplist.split(",")

        result = []
        for unit in list:
            result.append(unit.strip(" \t"))

        return result


if __name__ == "__main__":
    info = Info("OnStart", "etcdCluster", "etcd", "192.168.1.1, 192.168.1.2", "192.168.1.2", "192.168.1.2", "3", "")
    print(info.Backend, info.Iplist, info.Localip)
    print(info.GetIPListArray())
    print(info.CheckLocalIp())
    if info.GetServiceUrl(8500) != "192.168.1.2:8500":
        print("Error: info.GetServiceUrl(8500)!=192.168.1.2:8500")

    print("info.GetNodename():", info.GetNodename())
    if info.GetNodename() != "etcdCluster-192.168.1.2":
        print("Error: info.GetNodename() != Consul-192.168.1.2")

    print("info.GetBackendNodename():", info.GetBackendNodename())
    if info.GetBackendNodename() != "Etcd-192.168.1.2":
        print("Error: info.GetBackendNodename() != Consul-192.168.1.2")

    info = Info("OnStart", "etcd", "etcd", "192.168.1.1, 192.168.1.2", "192.168.1.5", "192.168.1.2", "3", "")
    print(info.Backend, info.Iplist, info.Localip)
    print(info.GetIPListArray())
    print(info.CheckLocalIp())
