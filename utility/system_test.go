package utility

import (
	"testing"
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
}