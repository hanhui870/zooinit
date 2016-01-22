package etcd

import (
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
)

type Api struct {
	conn client.KeysAPI
}

// Cluster no points need
// Stat can't use this method, Struct mismatch
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

func Context() (context.Context){
	return context.Background()
}