package etcd

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"

	"strings"
)

const (
	// 1s is too short
	DEFALUT_CLIENT_TIMEOUT           = 5 * time.Second
	DEFAULT_CLIENT_KEEPALIVE         = 30 * time.Second
	DEFALUT_CLIENT_TLS_SHAKE_TIMEOUT = 5 * time.Second
)

type ApiKeys struct {
	conn   client.KeysAPI
	client client.Client
}

type ApiMembers struct {
	conn   client.MembersAPI
	client client.Client
}

type ApiStats struct {
	client client.Client
}

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
		Endpoints: endpoints,
		Transport: transport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	return client.New(cfg)
}

// Stat can't use this method, Struct mismatch
func NewApiKeys(endpoints []string) (*ApiKeys, error) {
	c, err := NewClient(endpoints)
	if err != nil {
		return nil, err
	}

	return &ApiKeys{conn: client.NewKeysAPI(c), client: c}, nil
}

func (a *ApiKeys) Conn() client.KeysAPI {
	if a == nil {
		return nil
	}

	return a.conn
}

// Member use this method
func NewApiMember(endpoints []string) (*ApiMembers, error) {
	c, err := NewClient(endpoints)
	if err != nil {
		return nil, err
	}

	return &ApiMembers{conn: client.NewMembersAPI(c), client: c}, nil
}

func (a *ApiMembers) Conn() client.MembersAPI {
	if a == nil {
		return nil
	}

	return a.conn
}

func (a *ApiMembers) GetInitialClusterSetting() (config string, err error) {
	if a == nil {
		return "", errors.New("ApiMembers obj nil.")
	}

	list, err := a.Conn().List(Context())
	if err != nil {
		return "", err
	}

	var cfgs []string
	for _, unit := range list {
		if len(unit.PeerURLs) < 1 {
			continue
		}
		// may exist mutiple url
		for _, peer := range unit.PeerURLs {
			cfgs = append(cfgs, unit.Name+"="+peer)
		}
	}

	return strings.Join(cfgs, ","), nil
}

func (a *ApiMembers) GetInitialClusterEndpoints() (config []string, err error) {
	if a == nil {
		return nil, errors.New("ApiMembers obj nil.")
	}

	list, err := a.Conn().List(Context())
	if err != nil {
		return nil, err
	}

	var cfgs []string
	for _, unit := range list {
		if len(unit.PeerURLs) < 1 {
			continue
		}
		// may exist mutiple url
		for _, peer := range unit.PeerURLs {
			cfgs = append(cfgs, peer)
		}
	}

	return cfgs, nil
}

// Stat use this method
func NewApiStats(endpoints []string) (*ApiStats, error) {
	c, err := NewClient(endpoints)
	if err != nil {
		return nil, err
	}

	return &ApiStats{client: c}, nil
}

func Context() context.Context {
	return context.Background()
}
