package etcd

import (
	"testing"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"log"
	"time"
)

func TestOriginalEtcdClient(t *testing.T) {
	cfg := client.Config{
		Endpoints: []string{"http://registry.alishui.com:2379"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	kapi := client.NewKeysAPI(c)
	// set "/foo" key with "bar" value
	log.Print("Setting '/foo' key with 'bar' value")
	resp, err := kapi.Set(context.Background(), "/foo", "bar", nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Set is done. Metadata is %q\n", resp)
	}

	// get "/foo" key's value
	log.Print("Getting '/foo' key value")
	resp, err = kapi.Get(context.Background(), "/foo", nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Get is done. Metadata is %q\n", resp)
		// print value
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	}
}

func TestClusterApi(t *testing.T) {
	api, err := NewApiKeys([]string{"http://registry.alishui.com:2379"})
	if err != nil {
		t.Error("NewApi error:", err)
	}

	resp, err := api.Conn().Delete(context.Background(), "/haimi/service/discover", nil)
	if err != nil {
		//may not exist
		t.Log("Delete Error /haimi/service/discover:", err)
	} else {
		t.Logf("Delete /haimi/service/discover: %q", resp)
	}

	resp, err = api.Conn().Create(context.Background(), "/haimi/service/discover", "test,hahahahah")
	if err != nil {
		//Key already exists (/haimi/service/discover) [11]
		t.Error("Create Error /haimi/service/discover:", err)
	} else {
		t.Logf("Create /haimi/service/discover: %q", resp)
	}

	resp, err = api.Conn().Set(context.Background(), "/haimi/service/discover", "test,hahahahah", &client.SetOptions{TTL: 1 * time.Minute})
	if err != nil {
		//Key already exists (/haimi/service/discover) [11]
		t.Error("Set Error /haimi/service/discover:", err)
	} else {
		t.Logf("Set /haimi/service/discover: %q", resp)
	}

	resp, err = api.Conn().CreateInOrder(context.Background(), "/haimi/service/order", "test,hahahahah", &client.CreateInOrderOptions{TTL: 1 * time.Minute})
	if err != nil {
		//Key already exists (/haimi/service/discover) [11]
		t.Error("CreateInOrder Error /haimi/service/order:", err)
	} else {
		t.Logf("CreateInOrder /haimi/service/order: %q", resp)
	}

}

func TestErrorCodeClient(t *testing.T) {
	//type need to be identical with cast type.
	//err:=client.Error{Code:client.ErrorCodeNodeExist, Message:"node exist.", Index:34}
	err := &client.Error{Code: client.ErrorCodeNodeExist, Message: "node exist.", Index: 34}
	if !EqualEtcdError(err, client.ErrorCodeNodeExist) {
		t.Error("EqualEtcdError(err, client.ErrorCodeNodeExist) Found error.")
	}

	if EqualEtcdError(err, client.ErrorCodeEventIndexCleared) {
		t.Error("EqualEtcdError(err, client.ErrorCodeEventIndexCleared) Found error.")
	}
}
