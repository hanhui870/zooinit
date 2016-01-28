package etcd

import (
	"net/url"
	"net/http"
	"errors"
	"strconv"
	"encoding/json"
)

/*
/v2/stats/self
{
    "name": "default",
    "id": "ce2a822cea30bfca",
    "state": "StateLeader",
    "startTime": "2016-01-26T14:51:30.694261426+08:00",
    "leaderInfo": {
        "leader": "ce2a822cea30bfca",
        "uptime": "1h22m35.435654084s",
        "startTime": "2016-01-26T14:51:31.095511161+08:00"
    },
    "recvAppendRequestCnt": 0,
    "sendAppendRequestCnt": 0
}
*/
type StatSelf struct {
}

type StatLeader struct {
}

/*
/v2/stats/store
{
    "compareAndSwapFail": 0,
    "compareAndSwapSuccess": 0,
    "createFail": 0,
    "createSuccess": 2,
    "deleteFail": 0,
    "deleteSuccess": 0,
    "expireCount": 0,
    "getsFail": 4,
    "getsSuccess": 75,
    "setsFail": 2,
    "setsSuccess": 4,
    "updateFail": 0,
    "updateSuccess": 0,
    "watchers": 0
}
*/
type StatStore struct {
}


// client use offical etcd/client, has endpoints impl..
type memberAction struct {
	Prefix string
	Key    string
}

func NewMemberAction() (m *memberAction) {
	return &memberAction{Prefix:"v2", Key:"members"}
}


func (m *memberAction) HTTPRequest(url url.URL) *http.Request {
	if m == nil {
		return nil
	}

	req, _ := http.NewRequest("GET", url.String() + "/" + m.Prefix + "/" + m.Key, nil)

	return req
}

// A cluster Member
type Member struct {
	ClientURLs []string `json:"clientURLs"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	PeerURLs   []string `json:"peerURLs"`
}

type MemberList struct {
	Members []Member `json:"members"`
}

func unmarshalMembersResponse(code int, header http.Header, body []byte) (list *MemberList, err error) {
	switch code {
	case http.StatusOK:
		if len(body) == 0 {
			return nil, errors.New("Empty response body.")
		}

		var list MemberList
		err = json.Unmarshal(body, &list)

		if err != nil || list.Members == nil {
			return nil, errors.New("Error /v2/member parse:" + string(body) + " " + err.Error())
		}else {
			return &list, nil
		}

	default:
		return nil, errors.New("Http response code error: " + strconv.Itoa(code) + " " + string(body))
	}
}
