class Info(object):
    'Cluster Boot info'

    def __init__(self, service, iplist, localip):
        self.Service = service
        self.Iplist = iplist
        self.Localip = localip

    def GetIPListArray(self):
        list = self.Iplist.split(",")

        result = []
        for unit in list:
            result.append(unit.strip("  "))

        return result


if __name__ == "__main__":
    info = Info("etcd", "192.168.1.1, 192.168.1.2", "192.168.1.2")
    print(info.Service, info.Iplist, info.Localip)
    print(info.GetIPListArray())
