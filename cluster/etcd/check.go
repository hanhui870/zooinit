package etcd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"errors"
)

// An cluster server health check
type ServerCheck struct {
	Client string
	Peer   string
}

type health struct {
	//json need first char upcase
	Health string `json:"health"`
}

// Allow peer empty
func NewServerCheck(client, peer string) *ServerCheck {
	return &ServerCheck{Client: client, Peer: peer}
}

// check cluster is cluster
func (s *ServerCheck) IsHealth() bool {
	if s == nil {
		return false
	}

	isHealthy, _:=CheckHealth(s.Client)

	return isHealthy
}

// check cluster is cluster
func (h *health) IsHealth() bool {
	return &h != nil && h.Health == "true"
}

// check health need a max timeout 1s for quick fail
func CheckHealth(client string) (bool, error) {
	cli := &http.Client{Timeout: time.Second}
	resp, err := cli.Get(client + "/health")
	if err != nil {
		//cannot use nil as type bool in return argument
		return false, err
	}

	var heal health
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &heal)
	if err != nil {
		return false ,errors.New("Error /v2/json parse:" + string(body) +" "+ err.Error())
	}

	return heal.IsHealth(), nil
}
