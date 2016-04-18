// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
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

	isHealthy, err := CheckHealth("http://registry.alishui.com:2379")
	if isHealthy != true {
		t.Log("CheckHealth of http://registry.alishui.com:2379/ fasle, please check server is up.", err)
	} else {
		t.Log("CheckHealth of http://registry.alishui.com:2379/:", true)
	}
}
