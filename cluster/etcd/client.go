package etcd

import (
	"time"
	"net"
	"net/http"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"

)

const (
	DEFALUT_CLIENT_TIMEOUT = 1 * time.Second
	DEFAULT_CLIENT_KEEPALIVE = 30 * time.Second
	DEFALUT_CLIENT_TLS_SHAKE_TIMEOUT = 5 * time.Second
)


type Api struct {
	keyConn client.KeysAPI
	client client.Client
}

// Cluster no points need
// Stat can't use this method, Struct mismatch
func NewClient(endpoints []string) (client.Client, error) {
	var transport client.CancelableTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   DEFALUT_CLIENT_TIMEOUT,
			KeepAlive: DEFAULT_CLIENT_KEEPALIVE,
		}).Dial,
		TLSHandshakeTimeout: DEFALUT_CLIENT_TLS_SHAKE_TIMEOUT,
	}

	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               transport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	return client.New(cfg)
}

// Cluster no points need
// Stat can't use this method, Struct mismatch
func NewApi(endpoints []string) (*Api, error) {
	c, err := NewClient(endpoints)
	if err != nil {
		return nil, err
	}

	return &Api{keyConn:client.NewKeysAPI(c), client:c}, nil
}

func (a *Api) Conn() (client.KeysAPI) {
	if a == nil {
		return nil
	}

	return a.keyConn
}

func (a *Api) Members() (list *MemberList, err error) {
	if a == nil {
		return nil, nil
	}

	act:=NewMemberAction()

	resp, body, err := a.client.Do(Context(), act)

	if err != nil {
		return nil, err
	}

	return unmarshalMembersResponse(resp.StatusCode, resp.Header, body)
}

func Context() (context.Context) {
	return context.Background()
}