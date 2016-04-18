// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package cluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

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

func getLastestNodeIPList() ([]string, []string, error) {
	//flush last log info
	defer env.Logger.Sync()

	list, err := getLastestNodeList()
	if err != nil {
		return nil, nil, err
	}

	var nodeList []string
	var uuidList []string
	for _, nodeValue := range list {
		em, err := BuildElectionMemberFromJSON(nodeValue)
		if err != nil {
			env.Logger.Println("Fetch Invaild electioin node:", nodeValue, " will continue..")
			continue
		}
		nodeList = append(nodeList, em.Ip)
		uuidList = append(uuidList, em.Uuid)
	}

	nodeList = utility.RemoveDuplicateInOrder(nodeList)
	uuidList = utility.RemoveDuplicateInOrder(uuidList)
	if len(nodeList) != len(uuidList) {
		env.Logger.Fatalln("Unique UUID and Unique IP list length is not match, please check, IP:", nodeList, " UUID:", uuidList)
	}
	return nodeList, uuidList, nil
}

func getLastestNodeUUIDList() ([]string, error) {
	//flush last log info
	defer env.Logger.Sync()

	list, err := getLastestElectionMemberList()
	if err != nil {
		return nil, err
	}

	var nodeList []string
	for _, node := range list {
		nodeList = append(nodeList, node.Uuid)
	}

	nodeList = utility.RemoveDuplicateInOrder(nodeList)
	return nodeList, nil
}

func getLastestElectionMemberList() ([]*ElectionMember, error) {
	//flush last log info
	defer env.Logger.Sync()

	list, err := getLastestNodeList()
	if err != nil {
		return nil, err
	}

	var nodeList []*ElectionMember
	for _, nodeValue := range list {
		em, err := BuildElectionMemberFromJSON(nodeValue)
		if err != nil {
			env.Logger.Println("Fetch Invaild electioin node:", nodeValue, " will continue..")
			continue
		}
		nodeList = append(nodeList, em)
	}

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

func setClusterBootedUUIDs(uuidList []string) {
	//flush last log info
	defer env.Logger.Sync()

	if len(uuidList) < 1 {
		env.Logger.Fatalln("SetClusterBootedUUIDs len of uuidList is 0, please check.")
	}

	//master update booted uuid
	if env.UUID == uuidList[0] {
		env.Logger.Println("SetClusterBootedUUIDs The server uuid is first of uuidList, will update booted_uuids...")
		kvApi := getClientKeysApi()
		// Set booted uuid
		resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED_UUIDS, strings.Join(uuidList, ","), &client.SetOptions{PrevExist: client.PrevNoExist})
		if err != nil {
			env.Logger.Fatalln("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED_UUIDS+" error:", err)
		} else {
			env.Logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED_UUIDS+" ok, uuidList:", uuidList, resp)
		}

		// set boot uuid map
		buffer, err := json.Marshal(uuidMap)
		if err != nil {
			env.Logger.Fatalln("Create UUID map JSON error:", err)
		}
		resp, err = kvApi.Conn().Set(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED_UUID_MAP, string(buffer), &client.SetOptions{PrevExist: client.PrevNoExist})
		if err != nil {
			env.Logger.Fatalln("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED_UUID_MAP+" error:", err)
		} else {
			env.Logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED_UUID_MAP+" ok, JSON:", string(buffer), resp)
		}
	}
}

func getClusterBootedUUIDs() ([]string, error) {
	//flush last log info
	defer env.Logger.Sync()

	kvApi := getClientKeysApi()
	// Set booted uuid
	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED_UUIDS, &client.GetOptions{})
	if err != nil {
		return nil, errors.New("Etcd.Api() get " + CLUSTER_CONFIG_DIR_BOOTED_UUIDS + " error:" + err.Error())
	} else {
		return strings.Split(resp.Node.Value, ","), nil
	}
}
