// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package etcd

import (
	"testing"
)

func TestMembersTestApi(t *testing.T) {
	api, err := NewApiMember([]string{"http://registry.alishui.com:2379"})
	if err != nil {
		t.Error("NewApi error:", err)
	}

	list, err := api.Conn().List(Context())
	if err != nil {
		t.Error("Fetch members error:", err)
	} else {
		for _, value := range list {
			t.Log("Found Member:", value.Name, value.ClientURLs, value.PeerURLs, value.ID)
		}
	}

	cfg, err := api.GetInitialClusterSetting()
	if err != nil {
		t.Error("GetInitialClusterSetting error:", err)
	} else {
		t.Log("GetInitialClusterSetting:", cfg)
	}

	urls, err := api.GetInitialClusterEndpoints()
	if err != nil {
		t.Error("GetInitialClusterEndpoints error:", err)
	} else {
		t.Log("GetInitialClusterEndpoints:", urls)
	}
}
