package utility

import (
	"net"
	"testing"
)

func TestDNSSRVOriginal(t *testing.T) {
	if testing.Short() {
		t.Skip("TestDNSSRVOriginal skip for short.")
	}

	t.Logf("Test In net.LookupSRV('', '', _etcd._tcp.discovery.alishui.com) ")
	domain := "_etcd._tcp.discovery.alishui.com"
	cname, srvs, err := net.LookupSRV("", "", domain)
	if err != nil {
		t.Error("Fetch srv err:", err)
	} else {
		t.Log("cname:", cname, "srvs:")
		//dns_test.go:16: Index '\x00' value: value.Target 192.168.4.221., value.Port 2379, value.Weight 1, value.Priority 5
		for key, value := range srvs {
			t.Logf("Index %q value: value.Target %s, value.Port %d, value.Weight %d, value.Priority %d ParsedIP:", key, value.Target, value.Port, value.Weight, value.Priority, net.ParseIP(value.Target[:len(value.Target)-1]))
		}
	}

	t.Logf("Test In net.LookupSRV(etcd, tcp, discovery.alishui.com) ")
	domain = "discovery.alishui.com"
	cname, srvs, err = net.LookupSRV("etcd", "tcp", domain)
	if err != nil {
		t.Error("Fetch srv err:", err)
	} else {
		t.Log("cname:", cname, "srvs:")
		//dns_test.go:16: Index '\x00' value: value.Target 192.168.4.221., value.Port 2379, value.Weight 1, value.Priority 5
		for key, value := range srvs {
			t.Logf("Index %q value: value.Target %s, value.Port %d, value.Weight %d, value.Priority %d", key, value.Target, value.Port, value.Weight, value.Priority)
		}
	}
}

func TestDNSQueryNormal(t *testing.T) {
	if testing.Short() {
		t.Skip("TestDNSQueryNormal skip for short.")
	}

	res, err := net.LookupAddr("223.5.5.5")
	if err != nil {
		t.Error("net.LookupAddr(\"223.5.5.5\") error:", err)
	} else {
		t.Log("net.LookupAddr(\"223.5.5.5\") res:", res)
	}

	res, err = net.LookupHost("223.5.5.5")
	if err != nil {
		t.Error("net.LookupHost(\"223.5.5.5\") error:", err)
	} else {
		t.Log("net.LookupHost(\"223.5.5.5\") res:", res)
	}

	res, err = net.LookupHost("localhost")
	if err != nil {
		t.Error("net.LookupHost(\"localhost\") error:", err)
	} else {
		t.Log("net.LookupHost(\"localhost\") res:", res)
	}

	res, err = net.LookupHost("www.alishui.com")
	if err != nil {
		t.Error("net.LookupHost(\"www.alishui.com\") error:", err)
	} else {
		t.Log("net.LookupHost(\"www.alishui.com\") res:", res)
	}

	ip, err := net.LookupIP("www.alishui.com")
	if err != nil {
		t.Error("net.LookupIP(\"www.alishui.com\") error:", err)
	} else {
		t.Log("net.LookupIP(\"www.alishui.com\") res:", ip)
	}

	txt, err := net.LookupTXT("zjgsdx.com")
	if err != nil {
		t.Error("net.LookupIP(\"zjgsdx.com\") error:", err)
	} else {
		t.Log("net.LookupIP(\"zjgsdx.com\") res:", txt)
	}

	//find on the local machine
	port, err := net.LookupPort("tcp", "ssh")
	if err != nil {
		t.Error("net.LookupPort(\"localhost\") error:", err)
	} else {
		t.Log("net.LookupPort(\"localhost\") res:", port)
	}

}

func TestSRVService(t *testing.T) {
	srv, err := NewSRVServiceOfDomainAndService("xmpp-server", "tcp", "google.com")
	if err != nil {
		t.Error("NewSRVServiceOfDomainAndService err:", err)
	} else {
		t.Log("Get Random one:", srv.GetRankedRandomService())
	}
}
