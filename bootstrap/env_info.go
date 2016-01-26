package bootstrap

import (
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/go-ini/ini"
	"strings"

	loglocal "zooinit/log"
	"zooinit/utility"
)

// This basic discovery service bootstrap env info
type envInfo struct {
	//service name, also use for log
	service string

	// Bootstrap etcd cluster service for boot other cluster service.
	discoveryHost string
	discoveryPort string
	discoveryPeer string

	// Used for internal bootstrap for system, Only one member.
	internalHost    string
	internalPort    string
	internalPeer    string
	internalDataDir string
	internalWalDir  string

	// internal Boot PID
	// defalut 0
	internalCmdInstance *exec.Cmd

	// whether internalHost is the machine running program owns
	isSelfIp bool

	//localIP for boot
	localIP net.IP

	// cluster totol qurorum
	qurorum int
	// sec unit
	timeout time.Duration

	logPath string

	// Logger instance for service
	logger *log.Logger

	// boot command
	cmd        string
	cmdDataDir string
	cmdWalDir  string
}

// New env from file
func NewEnvInfoFile(fname string) *envInfo {
	iniobj := GetConfigInstance(fname)

	return NewEnvInfo(iniobj)
}

func NewEnvInfo(iniobj *ini.File) *envInfo {
	obj := new(envInfo)

	sec := iniobj.Section(CONFIG_SECTION)
	obj.service = sec.Key("service").String()
	if obj.service == "" {
		log.Fatalln("Config of service is empty.")
	}

	discovery := sec.Key("discovery").String()
	if discovery == "" {
		log.Fatalln("Config of discovery is empty.")
	}
	if strings.Count(discovery, ":") != 2 {
		log.Fatalln("Config of discovery need ip:port:peer format.")
	}
	obj.discoveryHost = discovery[0:strings.Index(discovery, ":")]
	obj.discoveryPort = discovery[strings.Index(discovery, ":")+1 : strings.LastIndex(discovery, ":")]
	obj.discoveryPeer = discovery[strings.LastIndex(discovery, ":")+1:]

	internal := sec.Key("internal").String()
	if internal == "" {
		log.Fatalln("Config of internal is empty.")
	}
	if strings.Count(internal, ":") != 1 {
		log.Fatalln("Config of internal need port:peer format.")
	}
	// Must be identical with discoveryHost
	obj.internalHost = obj.discoveryHost
	obj.internalPort = internal[0:strings.Index(internal, ":")]
	obj.internalPeer = internal[strings.LastIndex(internal, ":")+1:]

	path := sec.Key("internal.data.dir").String()
	if path == "" {
		log.Fatalln("Config of internal.data.dir is empty.")
	}
	obj.internalDataDir = path

	path = sec.Key("internal.wal.dir").String()
	if path == "" {
		log.Fatalln("Config of internal.wal.dir is empty.")
	}
	obj.internalWalDir = path

	qurorum, err := sec.Key("qurorum").Int()
	if err != nil {
		log.Fatalln("Config of qurorum is error:", err)
	}
	if qurorum < 3 {
		log.Fatalln("Config of qurorum must >=3")
	}
	obj.qurorum = qurorum

	timeout, err := sec.Key("timeout").Float64()
	if err != nil {
		log.Fatalln("Config of timeout is error:", err)
	}
	if timeout == 0 {
		obj.timeout = CLUSTER_BOOTSTRAP_TIMEOUT
	} else {
		obj.timeout = time.Duration(int(timeout * 1000000000))
	}

	obj.logPath = sec.Key("log.path").String()
	if obj.logPath == "" {
		log.Fatalln("Config of log.path is empty.")
	}

	obj.logger = loglocal.GetConsoleFileMultiLogger(loglocal.GenerateFileLogPathName(obj.logPath, obj.service))
	obj.logger.Println("Configure file parsed. Waiting to be boostrapped.")

	obj.cmd = sec.Key("boot.cmd").String()
	if obj.cmd == "" {
		log.Fatalln("Config of boot.cmd is empty.")
	}

	path = sec.Key("boot.data.dir").String()
	if path == "" {
		log.Fatalln("Config of boot.data.dir is empty.")
	}
	obj.cmdDataDir = path

	path = sec.Key("boot.wal.dir").String()
	if path == "" {
		log.Fatalln("Config of boot.wal.dir is empty.")
	}
	obj.cmdWalDir = path

	// Init Extra runtime info
	if utility.HasIpAddress(obj.internalHost) {
		obj.isSelfIp = true
		obj.localIP = net.ParseIP(obj.internalHost)
	} else {
		obj.isSelfIp = false

		localip, err := utility.GetLocalIPWithIntranet(obj.internalHost)
		if err != nil {
			obj.logger.Fatalln("utility.GetLocalIPWithIntranet Please check configuration of discovery is correct.")
		}
		obj.localIP = localip
	}

	return obj
}

// Fetch bootstrap command
func (e *envInfo) GetCmd() string {
	if e == nil {
		return ""
	}

	return e.cmd
}

func (e *envInfo) GetQurorum() int {
	if e == nil {
		return 0
	}

	return e.qurorum
}

func (e *envInfo) GetTimeout() time.Duration {
	if e == nil {
		return 0
	}

	return e.timeout
}

func (e *envInfo) Service() string {
	if e == nil {
		return ""
	}

	return e.service
}

func (e *envInfo) Logger() *log.Logger {
	if e == nil {
		return nil
	}

	return e.logger
}

func (e *envInfo) LocalIP() net.IP {
	if e == nil {
		return nil
	}

	return e.localIP
}

func (e *envInfo) GetDiscoveryHost() string {
	if e == nil {
		return ""
	}

	return e.discoveryHost
}

func (e *envInfo) GetDiscoveryPort() string {
	if e == nil {
		return ""
	}

	return e.discoveryPort
}
