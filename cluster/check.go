package cluster

import (
	"bytes"
	"encoding/json"
	"os"
	"time"
)

// Health check
type ServiceCheck interface {
	IsHealth() bool

	Members()

	AddMember() error

	DelMember() error
}

type ClusterMember struct {
	Name     string `json:"name"`
	Update   string `json:"update"`
	Localip  string `json:"localip"`
	Hostname string `json:"hostname"`
	State    bool   `json:"state"`
	Failed   int    `json:"failed"`
	UUID     string `json:"uuid"`
}

func NewClusterMember(env Env, State bool, Failed int) *ClusterMember {
	host, err := os.Hostname()
	if err != nil {
		host = env.GetLocalIP().String()
	}
	return &ClusterMember{Name: env.GetNodename(), Localip: env.GetLocalIP().String(), State: State, Failed: Failed, Update: time.Now().Format(time.RFC3339), Hostname: host, UUID: env.GetUUID()}
}

func (m *ClusterMember) ToJson() string {
	b, err := json.Marshal(m)
	if err == nil {
		return string(b)
	} else {
		return ""
	}
}

func (m *ClusterMember) IsHealth() bool {
	return m.State
}

func BuildCheckInfoFromJSON(str string) (*ClusterMember, error) {
	var member ClusterMember

	err := json.Unmarshal(bytes.NewBufferString(str).Bytes(), &member)
	if err != nil {
		return nil, err
	}

	return &member, nil
}
