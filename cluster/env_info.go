package cluster

import (
	"log"
	"time"

	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"

	"zooinit/config"
	loglocal "zooinit/log"
	"zooinit/utility"
)

// This cluster service bootstrap env info
type envInfo struct {
	BaseInfo

	// Cluster power backend
	clusterBackend string

	// Bootstrap etcd cluster service for boot other cluster service.
	discoveryMethod string
	discoveryTarget string
	discoveryPath   string

	// Ip hint use to found which ip for boot bind
	iphint string

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

	obj.Service = service
	if obj.Service == "" {
		log.Fatalln("Config of service is empty.")
	}

	// key for process now
	var keyNow string

	keyNow = "pid.path"
	obj.PidPath = config.GetValueString(keyNow, sec, c)
	if obj.PidPath == "" {
		log.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "log.channel"
	obj.LogChannel = config.GetValueString(keyNow, sec, c)
	if obj.LogChannel == "" || !utility.InSlice([]string{loglocal.LOG_FILE, loglocal.LOG_STDOUT, loglocal.LOG_MULTI}, obj.LogChannel) {
		log.Fatalln("Config of " + keyNow + " must be one of file/stdout/multi.")
	}

	keyNow = "log.path"
	obj.LogPath = config.GetValueString(keyNow, sec, c)
	if obj.LogPath == "" {
		log.Fatalln("Config of log.path is empty, log to stdout also need.")
	}

	obj.clusterBackend = backend

	// Construct logger instance
	if obj.LogChannel == "file" {
		obj.Logger = loglocal.GetFileLogger(loglocal.GenerateFileLogPathName(obj.LogPath, obj.Service))
	} else if obj.LogChannel == "stdout" {
		obj.Logger = loglocal.GetBufferedLogger()
	} else if obj.LogChannel == "multi" {
		obj.Logger = loglocal.GetConsoleFileMultiLogger(loglocal.GenerateFileLogPathName(obj.LogPath, obj.Service))
	}
	//flush last log info
	defer obj.Logger.Sync()

	obj.Logger.Println("Service name of cluster is:", obj.Service)

	keyNow = "discover.method"
	obj.discoveryMethod = config.GetValueString(keyNow, sec, c)
	if obj.discoveryMethod == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "discover.target"
	obj.discoveryTarget = config.GetValueString(keyNow, sec, c)
	if obj.discoveryTarget == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "discover.path"
	obj.discoveryPath = config.GetValueString(keyNow, sec, c)
	if obj.discoveryPath == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.discoveryPath = obj.discoveryPath + "/" + obj.Service

	keyNow = "qurorum"
	qurorum, err := config.GetValueInt(keyNow, sec, c)
	if err != nil {
		obj.Logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if qurorum < 3 {
		obj.Logger.Fatalln("Config of " + keyNow + " must >=3")
	}
	obj.Qurorum = qurorum

	keyNow = "timeout"
	timeout, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		obj.Logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if timeout == 0 {
		obj.Timeout = CLUSTER_BOOTSTRAP_TIMEOUT
	} else {
		obj.Timeout = time.Duration(int(timeout * 1000000000))
	}

	keyNow = "health.check.interval"
	checkInterval, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		obj.Logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if checkInterval > 60 || checkInterval < 1 {
		obj.Logger.Fatalln("Config of " + keyNow + " must be between 1-60 sec.")
	}
	if checkInterval == 0 {
		obj.HealthCheckInterval = CLUSTER_HEALTH_CHECK_INTERVAL
	} else {
		obj.HealthCheckInterval = time.Duration(int(checkInterval * 1000000000))
	}

	// Event process
	keyNow = "EVENT_ON_PRE_REGIST"
	obj.eventOnPreRegist = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPreRegist != "" {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnPreRegist)
	}
	keyNow = "EVENT_ON_POST_REGIST"
	obj.eventOnPostRegist = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPostRegist != "" {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnPostRegist)
	}
	keyNow = "EVENT_ON_REACH_QURORUM_NUM"
	obj.eventOnReachQurorumNum = config.GetValueString(keyNow, sec, c)
	if obj.eventOnReachQurorumNum != "" {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnReachQurorumNum)
	}
	keyNow = "EVENT_ON_PRE_START"
	obj.eventOnPreStart = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPreStart != "" {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnPreStart)
	}
	//required
	keyNow = "EVENT_ON_START"
	obj.eventOnStart = config.GetValueString(keyNow, sec, c)
	if obj.eventOnStart == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	} else {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnStart)
	}
	keyNow = "EVENT_ON_POST_START"
	obj.eventOnPostStart = config.GetValueString(keyNow, sec, c)
	if obj.eventOnPostStart == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	} else {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnPostStart)
	}
	keyNow = "EVENT_ON_CLUSTER_BOOTED"
	obj.eventOnClusterBooted = config.GetValueString(keyNow, sec, c)
	if obj.eventOnClusterBooted != "" {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnClusterBooted)
	}
	keyNow = "EVENT_ON_HEALTH_CHECK"
	obj.eventOnHealthCheck = config.GetValueString(keyNow, sec, c)
	if obj.eventOnHealthCheck == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	} else {
		obj.Logger.Println("Found event "+keyNow+":", obj.eventOnHealthCheck)
	}

	keyNow = "ip.hint"
	obj.iphint = config.GetValueString(keyNow, sec, c)
	if obj.iphint == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	// Find localip
	localip, err := utility.GetLocalIPWithIntranet(obj.iphint)
	if err != nil {
		obj.Logger.Fatalln("utility.GetLocalIPWithIntranet Please check configuration of discovery is correct.")
	}
	obj.LocalIP = localip
	obj.Logger.Println("Found localip for boot:", obj.LocalIP)

	// store app config, optional
	appSection := clusterSection + ".config"
	secApp, err := iniobj.GetSection(appSection)
	if err != nil {
		obj.Logger.Println("Config of app config section: " + appSection + " is not well configured, continue...")
	} else {
		obj.config = secApp.KeysHash()
		if len(obj.config) > 0 {
			obj.Logger.Println("Fetch app config section " + appSection + " KV values:")

			for key, value := range obj.config {
				obj.Logger.Println("Key:", key, " Value:", value)
			}
		} else {
			obj.Logger.Println("Fetch app config section: empty")
		}
	}

	obj.Logger.Println("Configure file parsed. Waiting to be boostrapped...")

	//create uuid
	obj.CreateUUID()

	return obj
}
