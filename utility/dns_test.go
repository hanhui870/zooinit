// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package utility

import (
	"math/rand"
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
		t.Error("net.LookupTXT(\"zjgsdx.com\") error:", err)
	} else {
		t.Log("net.LookupTXT(\"zjgsdx.com\") res:", txt)
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
	if testing.Short() {
		t.Skip("TestSRVService skip for short.")
	}

	var arr = []int{2}
	rt := rand.Intn(len(arr))
	if arr[rt] != 2 {
		t.Error("found rand.Intn(len(arr)) error.", rt)
	}

	srv, err := NewSRVServiceOfDomainAndService("etcd", "tcp", "discovery.alishui.com")
	if err != nil {
		t.Error("NewSRVServiceOfDomainAndService discovery.alishui.com err:", err)
	} else {
		t.Logf("result discovery.alishui.com: %q", srv)

		rs, err := srv.GetRankedService()
		if err != nil {
			t.Error("GetRankedRandomService err:", err)
		} else {
			t.Log("Get Ranked one:", rs.ip, rs.port)
		}

		//panic: runtime error: invalid memory address or nil pointer dereference [recovered] May not exist.
		//t.Log("Get Random one:", srv.GetRankedRandomService().cname)
		rs, err = srv.GetRandomService()
		if err != nil {
			t.Error("GetRandomService err:", err)
		} else {
			t.Log("Get Random one:", rs.ip, rs.port)
		}

		relist, err := srv.GetAllSrvList()
		if err != nil {
			t.Error("GetAllSrvList err:", err)
		} else {
			t.Log("Get GetAllSrvList endpoints:", relist.Endpoints())
		}
	}
}

func TestSRVInfoNormal(t *testing.T) {
	sv1, err := NewSRVInfoBuild("192.168.4.220", 2379)
	if err != nil {
		t.Error("NewSRVInfoBuild err:", err)
	}

	sv2, err := NewSRVInfoBuild("192.168.4.221", 2379)
	if err != nil {
		t.Error("NewSRVInfoBuild err:", err)
	}

	svlist := &SRVList{sv1, sv2}
	if svlist.Endpoints() != "192.168.4.220:2379,192.168.4.221:2379" {
		t.Error("svlist.Endpoints() err:", svlist.Endpoints())
	} else {
		t.Log("svlist.Endpoints() result:", svlist.Endpoints())
	}

}
