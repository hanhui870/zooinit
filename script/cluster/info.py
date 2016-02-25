class Info(object):
    'Cluster Boot info'

    def __init__(self, service, iplist, localip, masterip, qurorum):
        self.Service = service
        self.Iplist = iplist
        self.Localip = localip
        self.Masterip = masterip
        self.Qurorum = qurorum

    def GetIPListArray(self):
        list = self.Iplist.split(",")

        result = []
        for unit in list:
            result.append(unit.strip("  "))

        return result

    def CheckLocalIp(self):
        list = self.GetIPListArray()
        for ip in list:
            if ip == self.Localip:
                return True
        return False


if __name__ == "__main__":
    info = Info("etcd", "192.168.1.1, 192.168.1.2", "192.168.1.2", "192.168.1.2", "3")
    print(info.Service, info.Iplist, info.Localip)
    print(info.GetIPListArray())
    print(info.CheckLocalIp())

    info = Info("etcd", "192.168.1.1, 192.168.1.2", "192.168.1.5", "192.168.1.2", "3")
    print(info.Service, info.Iplist, info.Localip)
    print(info.GetIPListArray())
    print(info.CheckLocalIp())
