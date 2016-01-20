package utility

import (
	"testing"
	"net"
)

func TestFetchIPList(t *testing.T) {
	iplist, err := GetIpAddress("")
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range iplist {
		t.Log("Fetch for Interface:", key)
		for iter, ip := range value {
			t.Log("IP[", iter, "]=", ip.To4())
		}
	}

	t.Log("Has IP 192.168.4.1: ", HasIpAddress("192.168.4.1"))
	t.Log("Has IP 192.168.4.108: ", HasIpAddress("192.168.4.108"))
	t.Log("Has IP 127.0.0.1: ", HasIpAddress("127.0.0.1"))

	ip:=net.IPv4(192,168,4,108)
	t.Log("IP mask of 192.168.4.108: ", ip.Mask(net.IPv4Mask(255,255,255,0)))
	//actual
	actualMask:=MaskOFIpAddress("192.168.4.108")
	t.Log("IP mask of 192.168.4.108: ", actualMask.String())
	t.Log("Actual IP mask of 192.168.4.108: ", ip.Mask(actualMask))

	ip, err = GetLocalIPWithIntranet("192.168.4.199")
	if err!=nil {
		t.Error("GetLocalIPWithIntranet of 192.168.4.199:", err)
	}else{
		t.Log("Find the smae intranet of 192.168.4.199: ", ip)
	}

	ip, err = GetLocalIPWithIntranet("192.168.1.4")
	if err!=nil {
		t.Error("GetLocalIPWithIntranet of 192.168.1.4:", err)
	}else{
		t.Log("Find the smae intranet of 192.168.1.4: ", ip)
	}

}