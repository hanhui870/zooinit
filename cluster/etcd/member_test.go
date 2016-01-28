package etcd

import (
	"testing"
)

func TestMembersTestApi(t *testing.T) {
	api, err := NewApi([]string{"http://registry.alishui.com:2379"})
	if err != nil {
		t.Error("NewApi error:", err)
	}

	list, err:=api.Members()
	if err!=nil {
		t.Error("Fetch members error:", err)
	}else{
		for _, value := range list.Members {
			t.Logf("Found Member:", value.Name, value.ClientURLs, value.PeerURLs, value.Id)
		}
	}
}