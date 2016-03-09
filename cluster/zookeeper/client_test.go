package zookeeper

import (
	"testing"
	"time"

	"github.com/samuel/go-zookeeper/zk"

	"bytes"
)

const (
	ZK_HOST = "192.168.99.100:33064"
)

func TestCreateSignal(t *testing.T) {
	zkc, _, err := zk.Connect([]string{ZK_HOST}, time.Second)
	if err != nil {
		t.Fatalf("Connect returned error: %+v", err)
	}

	defer zkc.Close()

	path := "/gozktestsingle"

	if err := zkc.Delete(path, -1); err != nil && err != zk.ErrNoNode {
		t.Fatalf("Delete returned error: %+v", err)
	}
	if p, err := zkc.Create(path, []byte{1, 2, 3, 4}, 0, zk.WorldACL(zk.PermAll)); err != nil {
		t.Fatalf("Create returned error: %+v", err)
	} else if p != path {
		t.Fatalf("Create returned different path '%s' != '%s'", p, path)
	}
	if data, stat, err := zkc.Get(path); err != nil {
		t.Fatalf("Get returned error: %+v", err)
	} else if stat == nil {
		t.Fatal("Get returned nil stat")
	} else if len(data) < 4 {
		t.Fatal("Get returned wrong size data")
	}
}

func TestMulti(t *testing.T) {
	zkc, _, err := zk.Connect([]string{ZK_HOST}, time.Second)
	if err != nil {
		t.Fatalf("Connect returned error: %+v", err)
	}
	defer zkc.Close()

	path := "/gozktestmulti"

	if err := zkc.Delete(path, -1); err != nil && err != zk.ErrNoNode {
		t.Fatalf("Delete returned error: %+v", err)
	}
	ops := []interface{}{
		&zk.CreateRequest{Path: path, Data: []byte{1, 2, 3, 4}, Acl: zk.WorldACL(zk.PermAll)},
		&zk.SetDataRequest{Path: path, Data: bytes.NewBufferString("Hello world").Bytes(), Version: -1},
	}
	if res, err := zkc.Multi(ops...); err != nil {
		t.Fatalf("Multi returned error: %+v", err)
	} else if len(res) != 2 {
		t.Fatalf("Expected 2 responses got %d", len(res))
	} else {
		t.Logf("%+v", res)
	}
	if data, stat, err := zkc.Get(path); err != nil {
		t.Fatalf("Get returned error: %+v", err)
	} else if stat == nil {
		t.Fatal("Get returned nil stat")
	} else if len(data) < 4 {
		t.Fatal("Get returned wrong size data")
	}
}
