package cluster

import (
	"log"
	"net"
)

type Env interface {
	// get logger of underlevel app
	Logger() (*log.Logger)

	// localIP for boot
	LocalIP() (net.IP)

	// service name, also use for log
	Service() (string)
}
