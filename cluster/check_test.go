package cluster

import (
	"testing"
	"time"
)

func TestCheckServiceTests(t *testing.T) {
	ch := &ClusterMember{Name: "Consul-192.168.4.108", Localip: "192.168.4.108", Update: time.Now().Format(time.RFC3339), State: true}
	t.Log(ch.ToJson())

	chn := NewClusterMember("Consul-192.168.4.108", "192.168.4.108", true, 0)
	if ch.ToJson() != chn.ToJson() {
		t.Error("NewClusterMember Found error.")
	}

	if ch.IsHealth() != true {
		t.Error("Clustmember Health check failed.")
	}

	str := "{\"name\":\"zookeeper-192.168.4.221\",\"update\":\"2016-03-18T16:09:43+08:00\",\"localip\":\"192.168.4.221\",\"hostname\":\"t221.alishui.com\",\"state\":true,\"failed\":0}"
	unit, err := BuildFromJSON(str)
	if err != nil {
		t.Error("error parse BuildFromJSON")
	} else if unit.Localip != "192.168.4.221" {
		t.Error("error parse BuildFromJSON, unit.Localip error.")
	}
}
