package etcd

import (
	"time"

	"github.com/coreos/etcd/client"
)

type Api struct {
	conn client.KeysAPI
}

// Cluster no points need
func NewApi(endpoints []string) (*Api, error) {
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	return &Api{client.NewKeysAPI(c)}, nil
}

func (a *Api) Conn() (client.KeysAPI){
	if a==nil {
		return nil
	}

	return a.conn
}