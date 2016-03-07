package cluster

import (
	"testing"
	"time"
)

func TestCheckServiceTests(t *testing.T) {
	ch := &ClusterMember{Name: "Consul-192.168.4.108", Localip: "192.168.4.108", Update: time.Now().Format(time.RFC3339), State: true}
	t.Log(ch.ToJson())

	chn := NewClusterMember("Consul-192.168.4.108", "192.168.4.108", true)
	if ch.ToJson() != chn.ToJson() {
		t.Error("NewClusterMember Found error.")
	}

	if ch.IsHealth() != true {
		t.Error("Clustmember Health check failed.")
	}
}
