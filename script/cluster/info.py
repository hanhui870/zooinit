import json


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

    def GetUUIDMap(self):
        if self.UUIDMap is None:
            return dict()

        try:
            result = json.loads(self.UUIDMap)
        except Exception as err:
            result = []

        return result


if __name__ == "__main__":
    info = Info("OnStart", "etcdCluster", "etcd", "192.168.1.1, 192.168.1.2", "192.168.1.2", "192.168.1.2", "3",
                '{"uuu-dddd-1":"192.168.4.221","uuu-dddd-2":"192.168.4.222","uuu-dddd-3":"192.168.4.223","uuu-dddd-4":"192.168.4.224"}')
    print(info.Backend, info.Iplist, info.Localip)
    print(info.GetIPListArray())
    print(info.CheckLocalIp())
    if info.GetServiceUrl(8500) != "192.168.1.2:8500":
        print("Error: info.GetServiceUrl(8500)!=192.168.1.2:8500")

    print("info.GetNodename():", info.GetNodename())
    if info.GetNodename() != "etcdCluster-192.168.1.2":
        print("Error: info.GetNodename() != Consul-192.168.1.2")

    print(info.GetUUIDMap())
    if info.GetUUIDMap()["uuu-dddd-3"] != "192.168.4.223":
        print('Error: info.GetUUIDMap()["uuu-dddd-3"]!= "192.168.4.223"')

    info = Info("OnStart", "etcd", "etcd", "192.168.1.1, 192.168.1.2", "192.168.1.5", "192.168.1.2", "3")
    print(info.Backend, info.Iplist, info.Localip)
    print(info.GetIPListArray())
    print(info.CheckLocalIp())
