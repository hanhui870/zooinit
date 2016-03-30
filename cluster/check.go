package cluster

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/coreos/etcd/client"

	"zooinit/cluster/etcd"
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
	return &ClusterMember{Name: env.GetNodename(), Localip: env.GetLocalIP().String(), State: State, Failed: Failed, Update: time.Now().Format(time.RFC3339), Hostname: env.GetHostname(), UUID: env.GetUUID()}
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

// fetch /members list
func getLastestClusterMemberList() ([]*ClusterMember, error) {
	//flush last log info
	defer env.Logger.Sync()

	kvApi := getClientKeysApi()
	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_MEMBER_DIR, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, err
	}

	var memberList []*ClusterMember
	for _, node := range resp.Node.Nodes {
		if node.Dir {
			continue
		}

		memberInfo, err := BuildCheckInfoFromJSON(node.Value)
		if err != nil {
			env.Logger.Println("BuildCheckInfoFromJSON(node.Value) error:", err)
			continue
		}
		memberList = append(memberList, memberInfo)
	}
	return memberList, nil
}
