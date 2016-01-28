package etcd

import (
	"net/url"
	"net/http"
	"errors"
	"strconv"
	"encoding/json"
)

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

		if err != nil || list.Members==nil {
			return nil ,errors.New("Error /v2/member parse:" + string(body) +" "+ err.Error())
		}else{
			return &list, nil
		}

	default:
		return nil, errors.New("Http response code error: "+strconv.Itoa(code)+" "+string(body))
	}
}
