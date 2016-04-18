// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package etcd

import (
	"github.com/coreos/etcd/client"
	"testing"
	"time"
)

func TestWatchTestApi(t *testing.T) {
	api, err := NewApiKeys([]string{"http://registry.alishui.com:2379"})
	if err != nil {
		t.Error("NewApi error:", err)
	}

	go func() {
		time.Sleep(5 * time.Second)

		api.Conn().CreateInOrder(Context(), "/haimi/test/watcher/not_exist", "this is sequential node", &client.CreateInOrderOptions{})
	}()

	//could waiter a not exist node.
	w := api.Conn().Watcher("/haimi/test/watcher/not_exist", &client.WatcherOptions{Recursive: true}) //AfterIndex:56,如果有小于现有的,可以马上触发
	resp, err := w.Next(Context())

	if err != nil {
		t.Error("Fetch wathch error: ", err)
	} else {
		t.Logf("Fetch watch response: %q", resp)
	}

	time.Sleep(5 * time.Second)
}
