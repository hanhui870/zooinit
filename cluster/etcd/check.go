package etcd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type health struct {
	//json need first char upcase
	Health string `json:"health"`
}

// check cluster is cluster
func (h *health) IsHealth() bool {
	return &h != nil && h.Health == "true"
}

// check health need a max timeout 1s for quick fail
func CheckHealth(client string) bool {
	cli := &http.Client{Timeout: time.Second}
	resp, err := cli.Get(client + "/health")
	if err != nil {
		return false
	}

	var heal health
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &heal)
	if err != nil {
		println("Fetch :", string(body), err.Error())
		return false
	}

	return heal.IsHealth()
}
