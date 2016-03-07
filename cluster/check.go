package cluster

import (
	"encoding/json"
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
	Name    string `json:"name"`
	Update  string `json:"update"`
	Localip string `json:"localip"`
	State   bool   `json:"state"`
}

func NewClusterMember(Name, Localip string, State bool) *ClusterMember {
	return &ClusterMember{Name: Name, Localip: Localip, State: State, Update: time.Now().Format(time.RFC3339)}
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
