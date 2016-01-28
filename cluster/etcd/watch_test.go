package etcd

import (
	"testing"
)

func TestWatchTestApi(t *testing.T) {
	api, err := NewApiKeys([]string{"http://registry.alishui.com:2379"})
	if err != nil {
		t.Error("NewApi error:", err)
	}

	//could waiter a not exist node.
	w := api.Conn().Watcher("/haimi/test/watcher/not_exist", nil)
	resp, err := w.Next(Context())
}
