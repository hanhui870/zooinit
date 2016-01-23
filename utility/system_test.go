package utility

import (
	"testing"
	"net"
	"reflect"
	"strings"
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
	}else {
		t.Log("Find the smae intranet of 192.168.4.199: ", ip)
	}

	ip, err = GetLocalIPWithIntranet("192.168.1.4")
	if err != nil {
		t.Log("GetLocalIPWithIntranet of 192.168.1.4 Error:", err)
	}else {
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
		{in:"start", out:testout{path:"start", args:nil}},
		{in:"start -name hello", out:testout{path:"start", args:[]string{"-name", "hello"}}},
		// not support ' quote
		{in:"start -name 'hello", out:testout{path:"start", args:[]string{"-name", "'hello"}}},
		{in:"start -name \\'hello", out:testout{path:"start", args:[]string{"-name", "'hello"}}},
		{in:"start -name \\\"hello", out:testout{path:"start", args:[]string{"-name", "\"hello"}}},
		{in:"s\\ tart -name \\\"hello", out:testout{path:"s tart", args:[]string{"-name", "\"hello"}}},
		{in:"start -name \"hello\"", out:testout{path:"start", args:[]string{"-name", "hello"}}},
		{in:"start -name \"hello world\"", out:testout{path:"start", args:[]string{"-name", "hello world"}}},
	}

	for _, testCase := range testCases {
		path, args, err := ParseCmdStringWithParams(testCase.in)

		if err!=nil || path != testCase.out.path || !reflect.DeepEqual(args, testCase.out.args) {
			t.Error("Test failed for:", testCase,"path:", path, "result:", args)
		}else{
			t.Log("Test success for:", testCase,"path:", path, "result:", args)
		}
	}
}


func TestStringSlice(t *testing.T) {
	str := "TestParseCmdTestStringWithTestParams"

	var iter int
	find := "Test"
	//panic: runtime error: slice bounds out of range [recovered]
	//t.Log(find[5:])
	t.Log("String", str, "Slice based AT 5:", str[5:])

	start := 0;
	for {
		chunk := str[start:]
		iter = strings.Index(chunk, find)
		if iter == -1 {
			t.Log("Find no", find, "in:", chunk)
			break
		} else {
			t.Log("Find", find, "in:", chunk, "At", iter)
			//iter是基于当前start的
			start = start + iter + len(find)
			if start > len(str) {
				t.Log("Final break At:", start)
				break
			}

			t.Log("Next round start At:", start)
		}
	}

	str="fdsafda "
	t.Log("String trim of `fdsafda `:", strings.Trim(str, " fa"))
}
