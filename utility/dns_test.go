package utility

import (
	"testing"
	"net"
)

func TestDNSSRVOriginal(t *testing.T) {
	t.Logf("Test In net.LookupSRV('', '', _etcd._tcp.discovery.alishui.com) ")
	domain:="_etcd._tcp.discovery.alishui.com"
	cname, srvs, err:=net.LookupSRV("", "", domain)
	if err!=nil {
		t.Error("Fetch srv err:", err)
	}else{
		t.Log("cname:", cname, "srvs:")
		//dns_test.go:16: Index '\x00' value: value.Target 192.168.4.221., value.Port 2379, value.Weight 1, value.Priority 5
		for key, value := range srvs {
			t.Logf("Index %q value: value.Target %s, value.Port %d, value.Weight %d, value.Priority %d ParsedIP:", key, value.Target, value.Port, value.Weight, value.Priority, net.ParseIP(value.Target[:len(value.Target)-1]))
		}
	}

	t.Logf("Test In net.LookupSRV(etcd, tcp, discovery.alishui.com) ")
	domain="discovery.alishui.com"
	cname, srvs, err=net.LookupSRV("etcd", "tcp", domain)
	if err!=nil {
		t.Error("Fetch srv err:", err)
	}else{
		t.Log("cname:", cname, "srvs:")
		//dns_test.go:16: Index '\x00' value: value.Target 192.168.4.221., value.Port 2379, value.Weight 1, value.Priority 5
		for key, value := range srvs {
			t.Logf("Index %q value: value.Target %s, value.Port %d, value.Weight %d, value.Priority %d", key, value.Target, value.Port, value.Weight, value.Priority)
		}
	}
}
