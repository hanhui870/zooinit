package cluster

import (
	"net"
	loglocal "zooinit/log"
)

type Env interface {
	// get logger of underlevel app
	Logger() *loglocal.BufferedFileLogger

	// localIP for boot
	LocalIP() net.IP

	// service name, also use for log
	Service() string
}
