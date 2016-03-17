package bootstrap

import (
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"

	"zooinit/cluster"
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

	// localIP for boot
	localIP net.IP

	// cluster totol qurorum
	qurorum int
	// sec unit
	timeout time.Duration

	// Configuration of runtime log channel: file, write to file; stdout, write to stdout; multi, write both.
	logChannel string
	logPath    string

	// Logger instance for service
	logger *loglocal.BufferedFileLogger

	// boot command
	cmd          string
	cmdDataDir   string
	cmdWalDir    string
	cmdSnapCount int

	// Health check interval, default 2 sec, same to zookeeper ticktime.
	healthCheckInterval time.Duration
}

// New env from file
func NewEnvInfoFile(fname string) *envInfo {
	iniobj := config.GetConfigInstance(fname)

	return NewEnvInfo(iniobj, nil)
}

func NewEnvInfo(iniobj *ini.File, c *cli.Context) *envInfo {
	obj := new(envInfo)

	sec := iniobj.Section(CONFIG_SECTION)
	obj.service = sec.Key("service").String()
	if obj.service == "" {
		log.Fatalln("Config of service is empty.")
	}

	// key for process now
	var keyNow string

	keyNow = "log.channel"
	obj.logChannel = config.GetValueString(keyNow, sec, c)
	if obj.logChannel == "" || !utility.InSlice([]string{loglocal.LOG_FILE, loglocal.LOG_STDOUT, loglocal.LOG_MULTI}, obj.logChannel) {
		log.Fatalln("Config of " + keyNow + " must be one of file/stdout/multi.")
	}

	keyNow = "log.path"
	obj.logPath = config.GetValueString(keyNow, sec, c)
	if obj.logPath == "" {
		log.Fatalln("Config of " + keyNow + " is empty.")
	}

	// Construct logger instance
	if obj.logChannel == "file" {
		obj.logger = loglocal.GetFileLogger(loglocal.GenerateFileLogPathName(obj.logPath, obj.service))
	} else if obj.logChannel == "stdout" {
		obj.logger = loglocal.GetBufferedLogger()
	} else if obj.logChannel == "multi" {
		obj.logger = loglocal.GetConsoleFileMultiLogger(loglocal.GenerateFileLogPathName(obj.logPath, obj.service))
	}
	//flush last log info
	defer obj.logger.Sync()

	obj.logger.Println("Configure file parsed. Waiting to be boostrapped.")

	keyNow = "discovery"
	discovery := config.GetValueString(keyNow, sec, c)
	if discovery == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	if strings.Count(discovery, ":") != 2 {
		obj.logger.Fatalln("Config of " + keyNow + " need ip:port:peer format.")
	}
	obj.discoveryHost = discovery[0:strings.Index(discovery, ":")]
	obj.discoveryPort = discovery[strings.Index(discovery, ":")+1 : strings.LastIndex(discovery, ":")]
	obj.discoveryPeer = discovery[strings.LastIndex(discovery, ":")+1:]

	keyNow = "internal"
	internal := config.GetValueString(keyNow, sec, c)
	if internal == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	if strings.Count(internal, ":") != 1 {
		obj.logger.Fatalln("Config of " + keyNow + " need port:peer format.")
	}
	// Must be identical with discoveryHost
	obj.internalHost = obj.discoveryHost
	obj.internalPort = internal[0:strings.Index(internal, ":")]
	obj.internalPeer = internal[strings.LastIndex(internal, ":")+1:]

	keyNow = "internal.data.dir"
	path := config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.internalDataDir = path

	keyNow = "internal.wal.dir"
	path = config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.internalWalDir = path

	keyNow = "qurorum"
	qurorum, err := config.GetValueInt(keyNow, sec, c)
	if err != nil {
		obj.logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if qurorum < 3 {
		obj.logger.Fatalln("Config of " + keyNow + " must >=3")
	}
	obj.qurorum = qurorum

	keyNow = "timeout"
	timeout, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		obj.logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if timeout == 0 {
		obj.timeout = CLUSTER_BOOTSTRAP_TIMEOUT
	} else {
		obj.timeout = time.Duration(int(timeout * 1000000000))
	}

	keyNow = "health.check.interval"
	checkInterval, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		obj.logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if checkInterval > 60 || checkInterval < 1 {
		obj.logger.Fatalln("Config of " + keyNow + " must be between 1-60 sec.")
	}
	if checkInterval == 0 {
		obj.healthCheckInterval = cluster.CLUSTER_HEALTH_CHECK_INTERVAL
	} else {
		obj.healthCheckInterval = time.Duration(int(checkInterval * 1000000000))
	}

	keyNow = "boot.cmd"
	obj.cmd = config.GetValueString(keyNow, sec, c)
	if obj.cmd == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "boot.data.dir"
	path = config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.cmdDataDir = path

	keyNow = "boot.wal.dir"
	path = config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.cmdWalDir = path

	keyNow = "boot.snap.count"
	snapCount, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		obj.logger.Fatalln("Config of etcd "+keyNow+" is error:", err)
	}
	if snapCount > 100000 || snapCount < 100 {
		obj.logger.Fatalln("Config of etcd " + keyNow + " must between 100-100000")
	} else {
		obj.cmdSnapCount = int(snapCount)
	}

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

func (e *envInfo) GetInternalClientUrl() string {
	if e == nil {
		return ""
	}

	return "http://" + e.internalHost + ":" + e.internalPort
}

func (e *envInfo) GetInternalPeerUrl() string {
	if e == nil {
		return ""
	}

	return "http://" + e.internalHost + ":" + e.internalPeer
}

func (e *envInfo) GetClientUrl() string {
	if e == nil {
		return ""
	}

	return "http://" + env.localIP.String() + ":" + env.discoveryPort
}

func (e *envInfo) GetPeerUrl() string {
	if e == nil {
		return ""
	}

	return "http://" + env.localIP.String() + ":" + env.discoveryPeer
}

func (e *envInfo) GetNodename() string {
	if e == nil {
		return ""
	}

	return "Etcd-" + e.localIP.String()
}

func (e *envInfo) registerSignalWatch() {
	if e == nil {
		return
	}

	defer e.logger.Sync()

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
