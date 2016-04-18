// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package cluster

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestCheckServiceTests(t *testing.T) {
	host, err := os.Hostname()
	ch := &ClusterMember{Name: "Consul-192.168.4.108", Localip: "192.168.4.108", Update: time.Now().Format(time.RFC3339), State: true, Hostname: host}
	t.Log(ch.ToJson())

	env := &envInfo{clusterBackend: "Consul", BaseInfo: BaseInfo{Service: "Consul", LocalIP: net.ParseIP("192.168.4.108"), UUID: "uuid-test"}}
	chn := NewClusterMember(env, true, 0)
	t.Log("NewClusterMember Found error.", chn.ToJson())

	if ch.IsHealth() != true {
		t.Error("Clustmember Health check failed.")
	}

	str := "{\"name\":\"zookeeper-192.168.4.221\",\"update\":\"2016-03-18T16:09:43+08:00\",\"localip\":\"192.168.4.221\",\"hostname\":\"t221.alishui.com\",\"state\":true,\"failed\":0}"
	unit, err := BuildCheckInfoFromJSON(str)
	if err != nil {
		t.Error("error parse BuildFromJSON")
	} else if unit.Localip != "192.168.4.221" {
		t.Error("error parse BuildFromJSON, unit.Localip error.")
	}
	t.Log("Parsed value:", unit)
}
