package utility

import (
	"net"
	"reflect"
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

	ip := net.IPv4(192, 168, 4, 108)
	t.Log("IP mask of 192.168.4.108: ", ip.Mask(net.IPv4Mask(255, 255, 255, 0)))
	//actual
	actualMask := MaskOFIpAddress("192.168.4.108")
	t.Log("IP mask of 192.168.4.108: ", actualMask.String())
	t.Log("Actual IP mask of 192.168.4.108: ", ip.Mask(actualMask))

	ip, err = GetLocalIPWithIntranet("192.168.4.199")
	if err != nil {
		t.Log("GetLocalIPWithIntranet of 192.168.4.199 Error:", err)
	} else {
		t.Log("Find the smae intranet of 192.168.4.199: ", ip)
	}

	ip, err = GetLocalIPWithIntranet("192.168.1.4")
	if err != nil {
		t.Log("GetLocalIPWithIntranet of 192.168.1.4 Error:", err)
	} else {
		t.Log("Find the smae intranet of 192.168.1.4: ", ip)
	}

}

func TestParseCmdStringWithParams(t *testing.T) {
	type testout struct {
		path string
		args []string
	}
	type tests struct {
		in  string
		out testout
	}

	testCases := []tests{
		{in: "start", out: testout{path: "start", args: nil}},
		{in: "start -name hello", out: testout{path: "start", args: []string{"-name", "hello"}}},
		// not support ' quote
		{in: "start -name 'hello", out: testout{path: "start", args: []string{"-name", "'hello"}}},
		{in: "start -name \\'hello", out: testout{path: "start", args: []string{"-name", "'hello"}}},
		{in: "start -name \\\"hello", out: testout{path: "start", args: []string{"-name", "\"hello"}}},
		{in: "s\\ tart -name \\\"hello", out: testout{path: "s tart", args: []string{"-name", "\"hello"}}},
		{in: "start -name \"hello\"", out: testout{path: "start", args: []string{"-name", "hello"}}},
		{in: "start -name \"hello world\"", out: testout{path: "start", args: []string{"-name", "hello world"}}},
		//201602.25: fail consul agent -server -data-dir=\"/tmp/consul\" -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108
		{in: "consul agent -server -data-dir=/tmp/consul -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108",
			out: testout{path: "consul", args: []string{"agent", "-server", "-data-dir=/tmp/consul", "-bootstrap-expect", "3", "-bind=192.168.4.108", "-client=192.168.4.108"}}},
	}

	for _, testCase := range testCases {
		path, args, err := ParseCmdStringWithParams(testCase.in)

		if err != nil || path != testCase.out.path || !reflect.DeepEqual(args, testCase.out.args) {
			t.Error("Test failed for:", testCase.in, "path:", path, "result:", args, "expect:", testCase.out.path, testCase.out.args)
		} else {
			t.Log("Test success for:", testCase.in, "path:", path, "result:", args)
		}
	}
}
