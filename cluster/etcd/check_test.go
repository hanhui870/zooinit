package etcd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCheckServiceClient(t *testing.T) {
	buf := bytes.NewBufferString(`{"health": "true"}`)
	var heal health
	err := json.Unmarshal(buf.Bytes(), &heal)
	if err != nil {
		t.Errorf("json.Unmarshal(heal, &heal) err:", err)
	} else if heal.IsHealth() != true {
		t.Errorf("heal.IsHealth()==true Error, %b, &s", heal.IsHealth(), heal.Health)
	} else {
		t.Log("heal.IsHealth()=", heal.IsHealth())
	}

	if CheckHealth("http://localhost:2379") != true {
		t.Log("CheckHealth of http://localhost:2379/ fasle, please check server is up.")
	} else {
		t.Log("CheckHealth of http://localhost:2379/:", true)
	}
}
