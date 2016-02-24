package bootstrap

import (
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-ini/ini"

	"zooinit/config"
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
	logger *loglocal.BufferedFileLogger

	// boot command
	cmd        string
	cmdDataDir string
	cmdWalDir  string
}

// New env from file
func NewEnvInfoFile(fname string) *envInfo {
	iniobj := config.GetConfigInstance(fname)

	return NewEnvInfo(iniobj)
}

func NewEnvInfo(iniobj *ini.File) *envInfo {
	obj := new(envInfo)

	sec := iniobj.Section(CONFIG_SECTION)
	obj.service = sec.Key("service").String()
	if obj.service == "" {
		log.Fatalln("Config of service is empty.")
	}

	obj.logPath = sec.Key("log.path").String()
	if obj.logPath == "" {
		log.Fatalln("Config of log.path is empty.")
	}

	obj.logger = loglocal.GetConsoleFileMultiLogger(loglocal.GenerateFileLogPathName(obj.logPath, obj.service))
	//flush last log info
	defer obj.logger.Sync()

	//register signal watcher
	obj.registerSignalWatch()

	obj.logger.Println("Configure file parsed. Waiting to be boostrapped.")

	discovery := sec.Key("discovery").String()
	if discovery == "" {
		obj.logger.Fatalln("Config of discovery is empty.")
	}
	if strings.Count(discovery, ":") != 2 {
		obj.logger.Fatalln("Config of discovery need ip:port:peer format.")
	}
	obj.discoveryHost = discovery[0:strings.Index(discovery, ":")]
	obj.discoveryPort = discovery[strings.Index(discovery, ":")+1 : strings.LastIndex(discovery, ":")]
	obj.discoveryPeer = discovery[strings.LastIndex(discovery, ":")+1:]

	internal := sec.Key("internal").String()
	if internal == "" {
		obj.logger.Fatalln("Config of internal is empty.")
	}
	if strings.Count(internal, ":") != 1 {
		obj.logger.Fatalln("Config of internal need port:peer format.")
	}
	// Must be identical with discoveryHost
	obj.internalHost = obj.discoveryHost
	obj.internalPort = internal[0:strings.Index(internal, ":")]
	obj.internalPeer = internal[strings.LastIndex(internal, ":")+1:]

	path := sec.Key("internal.data.dir").String()
	if path == "" {
		obj.logger.Fatalln("Config of internal.data.dir is empty.")
	}
	obj.internalDataDir = path

	path = sec.Key("internal.wal.dir").String()
	if path == "" {
		obj.logger.Fatalln("Config of internal.wal.dir is empty.")
	}
	obj.internalWalDir = path

	qurorum, err := sec.Key("qurorum").Int()
	if err != nil {
		obj.logger.Fatalln("Config of qurorum is error:", err)
	}
	if qurorum < 3 {
		obj.logger.Fatalln("Config of qurorum must >=3")
	}
	obj.qurorum = qurorum

	timeout, err := sec.Key("timeout").Float64()
	if err != nil {
		obj.logger.Fatalln("Config of timeout is error:", err)
	}
	if timeout == 0 {
		obj.timeout = CLUSTER_BOOTSTRAP_TIMEOUT
	} else {
		obj.timeout = time.Duration(int(timeout * 1000000000))
	}

	obj.cmd = sec.Key("boot.cmd").String()
	if obj.cmd == "" {
		obj.logger.Fatalln("Config of boot.cmd is empty.")
	}

	path = sec.Key("boot.data.dir").String()
	if path == "" {
		obj.logger.Fatalln("Config of boot.data.dir is empty.")
	}
	obj.cmdDataDir = path

	path = sec.Key("boot.wal.dir").String()
	if path == "" {
		obj.logger.Fatalln("Config of boot.wal.dir is empty.")
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

func (e *envInfo) Logger() *loglocal.BufferedFileLogger {
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

func (e *envInfo) registerSignalWatch() {
	if e == nil {
		return
	}

	sg := utility.NewSignalCatcher()
	call := utility.NewSignalCallback(func(sig os.Signal, data interface{}) {
		e.logger.Println("Receive signal: " + sig.String() + " App will terminate, bye.")
		e.logger.Sync()
	}, nil)

	sg.SetDefault(call)
	sg.EnableExit()
	e.logger.Println("Init System SignalWatcher, catch list:", strings.Join(sg.GetSignalStringList(), ", "))

	sg.RegisterAndServe()
}
