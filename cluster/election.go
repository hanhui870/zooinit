package cluster

import (
	"bytes"
	"encoding/json"

	"github.com/coreos/etcd/client"

	"zooinit/cluster/etcd"
	"zooinit/utility"
)

type ElectionMember struct {
	Ip       string `json:"ip"`
	Uuid     string `json:"uuid"`
	Hostname string `json:"hostname"`
}

func NewElectionMemberEnv(env Env) *ElectionMember {
	return &ElectionMember{Uuid: env.GetUUID(), Ip: env.GetLocalIP().String(), Hostname: env.GetHostname()}
}

func NewElectionMember(ip, uuid, host string) *ElectionMember {
	return &ElectionMember{Uuid: uuid, Ip: ip, Hostname: host}
}

func (m *ElectionMember) ToJson() string {
	b, err := json.Marshal(m)
	if err == nil {
		return string(b)
	} else {
		return ""
	}
}

func BuildElectionMemberFromJSON(str string) (*ElectionMember, error) {
	var member ElectionMember

	err := json.Unmarshal(bytes.NewBufferString(str).Bytes(), &member)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func getLastestNodeIPList() ([]string, error) {
	//flush last log info
	defer env.Logger.Sync()

	list, err := getLastestNodeList()
	if err != nil {
		return nil, err
	}

	var nodeList []string
	for _, nodeValue := range list {
		em, err := BuildElectionMemberFromJSON(nodeValue)
		if err != nil {
			env.Logger.Println("Fetch Invaild electioin node:", nodeValue, " will continue..")
			continue
		}
		nodeList = append(nodeList, em.Ip)
	}

	nodeList = utility.RemoveDuplicateInOrder(nodeList)
	return nodeList, nil
}

func getLastestNodeUUIDList() ([]string, error) {
	//flush last log info
	defer env.Logger.Sync()

	list, err := getLastestNodeList()
	if err != nil {
		return nil, err
	}

	var nodeList []string
	for _, nodeValue := range list {
		em, err := BuildElectionMemberFromJSON(nodeValue)
		if err != nil {
			env.Logger.Println("Fetch Invaild electioin node:", nodeValue, " will continue..")
			continue
		}
		nodeList = append(nodeList, em.Uuid)
	}

	nodeList = utility.RemoveDuplicateInOrder(nodeList)
	return nodeList, nil
}

func getLastestNodeList() ([]string, error) {
	//flush last log info
	defer env.Logger.Sync()

	kvApi := getClientKeysApi()

	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_ELECTION_DIR, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, err
	} else {
		var nodeList []string
		for _, node := range resp.Node.Nodes {
			if node.Dir || !etcd.CheckInOrderKeyFormat(node.Key) {
				continue
			}
			nodeList = append(nodeList, node.Value)
		}

		nodeList = utility.RemoveDuplicateInOrder(nodeList)
		return nodeList, nil
	}
}
