package cluster

import (
	"log"
	"time"

	"github.com/go-ini/ini"

	"zooinit/config"
	loglocal "zooinit/log"
)

// This cluster service bootstrap env info
type envInfo struct {
	//service name, also use for log
	service string

	// Bootstrap etcd cluster service for boot other cluster service.
	discoveryMethod string
	discoveryTarget string
	discoveryPath   string

	// cluster totol qurorum
	qurorum int
	// sec unit
	timeout time.Duration

	logPath string

	// Logger instance for service
	logger *log.Logger

	// boot event related
	eventOnStart string

	// app start up configuration, app can fetch through env variables
	config map[string]string
}

// New env from file
func NewEnvInfoFile(fname string, cluster string) *envInfo {
	iniobj := config.GetConfigInstance(fname)

	return NewEnvInfo(iniobj, cluster)
}

func NewEnvInfo(iniobj *ini.File, cluster string) *envInfo {
	obj := new(envInfo)

	// init map
	obj.config = make(map[string]string)

	clusterSection := CONFIG_SECTION + "." + cluster
	sec, err := iniobj.GetSection(clusterSection)
	if err != nil {
		log.Fatalln("Config of section: " + clusterSection + " is not well configured.")
	}
	obj.service = sec.Key("service").String()
	if obj.service == "" {
		log.Fatalln("Config of service is empty.")
	}

	obj.logPath = sec.Key("log.path").String()
	if obj.logPath == "" {
		log.Fatalln("Config of log.path is empty.")
	}
	obj.logger = loglocal.GetConsoleFileMultiLogger(loglocal.GenerateFileLogPathName(obj.logPath, obj.service))
	obj.logger.Println("Configure file parsed. Waiting to be boostrapped.")

	obj.discoveryMethod = sec.Key("discover.method").String()
	if obj.discoveryMethod == "" {
		obj.logger.Fatalln("Config of discover.method is empty.")
	}

	obj.discoveryTarget = sec.Key("discover.target").String()
	if obj.discoveryTarget == "" {
		obj.logger.Fatalln("Config of discover.target is empty.")
	}

	obj.discoveryPath = sec.Key("discover.path").String()
	if obj.discoveryPath == "" {
		obj.logger.Fatalln("Config of discover.path is empty.")
	}

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

	// Event process
	obj.discoveryTarget = sec.Key("event.OnStart").String()
	if obj.discoveryTarget == "" {
		obj.logger.Fatalln("Config of event.OnStart is empty.")
	}

	// store app config
	appSection := clusterSection + ".config"
	secApp, err := iniobj.GetSection(appSection)
	if err != nil {
		obj.logger.Fatalln("Config of app config section: " + appSection + " is well configured.")
	}
	obj.config = secApp.KeysHash()
	if len(obj.config) > 0 {
		obj.logger.Println("Fetch app config section " + appSection + " KV values:")

		for key, value := range obj.config {
			obj.logger.Println("Key:", key, " Value:", value)
		}
	} else {
		obj.logger.Println("Fetch app config section: empty")
	}

	return obj
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
