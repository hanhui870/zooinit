package cluster

import (
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"

	"zooinit/config"
	loglocal "zooinit/log"
	"zooinit/utility"
)

// This cluster service bootstrap env info
type envInfo struct {
	// service name, also use for log
	service string
	// Cluster power backend
	clusterBackend string

	// Bootstrap etcd cluster service for boot other cluster service.
	discoveryMethod string
	discoveryTarget string
	discoveryPath   string

	// cluster totol qurorum
	qurorum int
	// sec unit
	timeout time.Duration

	// Configuration of runtime log channel: file, write to file; stdout, write to stdout; multi, write both.
	logChannel string
	logPath    string

	// Logger instance for service
	logger *loglocal.BufferedFileLogger

	// localIP for boot
	localIP net.IP
	// Ip hint use to found which ip for boot bind
	iphint string

	// Health check interval, default 2 sec, same to zookeeper ticktime.
	healthCheckInterval time.Duration

	// boot event related
	eventOnPreRegist       string
	eventOnPostRegist      string
	eventOnReachQurorumNum string
	eventOnPreStart        string
	eventOnStart           string
	eventOnPostStart       string
	eventOnClusterBooted   string
	eventOnHealthCheck     string

	// app start up configuration, app can fetch through env variables
	config map[string]string
}

// New env from file
func NewEnvInfoFile(fname string, backend, service string) *envInfo {
	iniobj := config.GetConfigInstance(fname)

	return NewEnvInfo(iniobj, backend, service, nil)
}

func NewEnvInfo(iniobj *ini.File, backend, service string, c *cli.Context) *envInfo {
	obj := new(envInfo)

	// init map
	obj.config = make(map[string]string)

	clusterSection := CONFIG_SECTION + "." + backend
	sec, err := iniobj.GetSection(clusterSection)
	if err != nil {
		log.Fatalln("Config of section: " + clusterSection + " is not well configured.")
	}

	obj.service = service
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
		log.Fatalln("Config of log.path is empty, log to stdout also need.")
	}

	obj.clusterBackend = backend

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

	obj.logger.Println("Service name of cluster is:", obj.service)

	keyNow = "discover.method"
	obj.discoveryMethod = config.GetValueString(keyNow, sec, c)
	if obj.discoveryMethod == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "discover.target"
	obj.discoveryTarget = config.GetValueString(keyNow, sec, c)
	if obj.discoveryTarget == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "discover.path"
	obj.discoveryPath = config.GetValueString(keyNow, sec, c)
	if obj.discoveryPath == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.discoveryPath = obj.discoveryPath + "/" + obj.service

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
		obj.healthCheckInterval = CLUSTER_HEALTH_CHECK_INTERVAL
	} else {
		obj.healthCheckInterval = time.Duration(int(checkInterval * 1000000000))
	}

	// Event process
	keyNow = "EVENT_ON_PRE_REGIST"
	obj.eventOnPreRegist = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPreRegist != "" {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnPreRegist)
	}
	keyNow = "EVENT_ON_POST_REGIST"
	obj.eventOnPostRegist = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPostRegist != "" {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnPostRegist)
	}
	keyNow = "EVENT_ON_REACH_QURORUM_NUM"
	obj.eventOnReachQurorumNum = config.GetValueString(keyNow, sec, c)
	if obj.eventOnReachQurorumNum != "" {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnReachQurorumNum)
	}
	keyNow = "EVENT_ON_PRE_START"
	obj.eventOnPreStart = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPreStart != "" {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnPreStart)
	}
	//required
	keyNow = "EVENT_ON_START"
	obj.eventOnStart = config.GetValueString(keyNow, sec, c)
	if obj.eventOnStart == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	} else {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnStart)
	}
	keyNow = "EVENT_ON_POST_START"
	obj.eventOnPostStart = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPostStart == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	} else {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnPostStart)
	}
	keyNow = "EVENT_ON_CLUSTER_BOOTED"
	obj.eventOnClusterBooted = config.GetValueString(keyNow, sec, c)
	if obj.eventOnClusterBooted != "" {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnClusterBooted)
	}
	keyNow = "EVENT_ON_HEALTH_CHECK"
	obj.eventOnHealthCheck = config.GetValueString(keyNow, sec, c)
	if obj.eventOnHealthCheck == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	} else {
		obj.logger.Println("Found event "+keyNow+":", obj.eventOnHealthCheck)
	}

	keyNow = "ip.hint"
	obj.iphint = config.GetValueString(keyNow, sec, c)
	if obj.iphint == "" {
		obj.logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	// Find localip
	localip, err := utility.GetLocalIPWithIntranet(obj.iphint)
	if err != nil {
		obj.logger.Fatalln("utility.GetLocalIPWithIntranet Please check configuration of discovery is correct.")
	}
	obj.localIP = localip
	obj.logger.Println("Found localip for boot:", obj.localIP)

	// store app config, optional
	appSection := clusterSection + ".config"
	secApp, err := iniobj.GetSection(appSection)
	if err != nil {
		obj.logger.Println("Config of app config section: " + appSection + " is not well configured, continue...")
	} else {
		obj.config = secApp.KeysHash()
		if len(obj.config) > 0 {
			obj.logger.Println("Fetch app config section " + appSection + " KV values:")

			for key, value := range obj.config {
				obj.logger.Println("Key:", key, " Value:", value)
			}
		} else {
			obj.logger.Println("Fetch app config section: empty")
		}
	}

	obj.logger.Println("Configure file parsed. Waiting to be boostrapped...")

	return obj
}

func (e *envInfo) GetQurorum() int {
	return e.qurorum
}

func (e *envInfo) GetTimeout() time.Duration {
	return e.timeout
}

func (e *envInfo) Service() string {
	if e == nil {
		return ""
	}

	return e.service
}

func (e *envInfo) Logger() *loglocal.BufferedFileLogger {
	return e.logger
}

func (e *envInfo) GetNodename() string {
	return e.clusterBackend + "-" + e.localIP.String()
}

func (e *envInfo) registerSignalWatch() {
	defer e.logger.Sync()

	sg := utility.NewSignalCatcher()
	stack := utility.NewSignalCallStack()
	sg.SetDefault(stack)
	sg.EnableExit()

	call := utility.NewSignalCallback(func(sig os.Signal, data interface{}) {
		defer e.logger.Sync()
		e.logger.Println("Receive signal: " + sig.String() + " App will terminate, bye.")
	}, nil)
	stack.Add(call)

	e.logger.Println("Init System SignalWatcher, catch list:", strings.Join(sg.GetSignalStringList(), ", "))

	sg.RegisterAndServe()
}
