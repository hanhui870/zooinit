package cluster

import (
	"bytes"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"
	"github.com/twinj/uuid"

	"io/ioutil"
	"zooinit/config"
	loglocal "zooinit/log"
	"zooinit/utility"
)

type Env interface {
	// get logger of underlevel app
	GetLogger() *loglocal.BufferedFileLogger

	// localIP for boot
	GetLocalIP() net.IP

	// service name, also use for log
	GetService() string

	// Get Pid file path
	GetPidPath() string

	GetNodename() string

	GetQurorum() int

	GetUUID() string

	GetHostname() string
}

type BaseInfo struct {
	// uuid of service
	UUID string

	// service name, also use for log
	Service string

	//Pid file path
	PidPath string

	// cluster totol qurorum
	Qurorum int
	// sec unit
	Timeout time.Duration

	// localIP for boot
	LocalIP net.IP
	// Ip hint use to found which ip for boot bind
	iphint string

	// Configuration of runtime log channel: file, write to file; stdout, write to stdout; multi, write both.
	LogChannel string
	LogPath    string

	// Health check interval, default 2 sec, same to zookeeper ticktime.
	HealthCheckInterval time.Duration

	// Term Signal catcher
	Sc *utility.SignalCatcher

	// Logger instance for service
	Logger *loglocal.BufferedFileLogger
}

func (e *BaseInfo) ParseConfigFile(sec *ini.Section, c *cli.Context) {
	// key for process now
	var keyNow string

	keyNow = "log.channel"
	e.LogChannel = config.GetValueString(keyNow, sec, c)
	if e.LogChannel == "" || !utility.InSlice([]string{loglocal.LOG_FILE, loglocal.LOG_STDOUT, loglocal.LOG_MULTI}, e.LogChannel) {
		log.Fatalln("Config of " + keyNow + " must be one of file/stdout/multi.")
	}

	keyNow = "log.path"
	e.LogPath = config.GetValueString(keyNow, sec, c)
	if e.LogPath == "" {
		log.Fatalln("Config of " + keyNow + " is empty.")
	}

	// Construct logger instance
	if e.LogChannel == "file" {
		e.Logger = loglocal.GetFileLogger(loglocal.GenerateFileLogPathName(e.LogPath, e.Service))
	} else if e.LogChannel == "stdout" {
		e.Logger = loglocal.GetBufferedLogger()
	} else if e.LogChannel == "multi" {
		e.Logger = loglocal.GetConsoleFileMultiLogger(loglocal.GenerateFileLogPathName(e.LogPath, e.Service))
	}

	//flush last log info
	defer e.Logger.Sync()

	keyNow = "pid.path"
	e.PidPath = config.GetValueString(keyNow, sec, c)
	if e.PidPath == "" {
		e.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "qurorum"
	qurorum, err := config.GetValueInt(keyNow, sec, c)
	if err != nil {
		e.Logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if qurorum < 3 {
		e.Logger.Fatalln("Config of " + keyNow + " must >=3")
	}
	e.Qurorum = qurorum

	keyNow = "timeout"
	timeout, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		e.Logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if timeout == 0 {
		e.Timeout = CLUSTER_BOOTSTRAP_TIMEOUT
	} else {
		e.Timeout = time.Duration(int(timeout * 1000000000))
	}

	keyNow = "health.check.interval"
	checkInterval, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		e.Logger.Fatalln("Config of "+keyNow+" is error:", err)
	}
	if checkInterval > 60 || checkInterval < 1 {
		e.Logger.Fatalln("Config of " + keyNow + " must be between 1-60 sec.")
	}
	if checkInterval == 0 {
		e.HealthCheckInterval = CLUSTER_HEALTH_CHECK_INTERVAL
	} else {
		e.HealthCheckInterval = time.Duration(int(checkInterval * 1000000000))
	}

	keyNow = "ip.hint"
	e.iphint = config.GetValueString(keyNow, sec, c)
	if e.iphint == "" {
		e.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	// Find localip
	localip, err := utility.GetLocalIPWithIntranet(e.iphint)
	if err != nil {
		e.Logger.Fatalln("utility.GetLocalIPWithIntranet Please check configuration of discovery is correct.")
	}
	e.LocalIP = localip
	e.Logger.Println("Found localip for boot:", e.LocalIP)
}

func (e *BaseInfo) CreateUUID() string {
	defer e.GetLogger().Sync()

	ufile := e.GetPidPath() + "/" + e.GetService() + ".uuid"

	e.GetLogger().Println("Read and check uuid:", ufile)

	//pid create dir if needed
	err := os.MkdirAll(filepath.Dir(ufile), loglocal.DEFAULT_LOGDIR_MODE)
	if err != nil {
		e.GetLogger().Fatalln("Create pid dir error, info:", err)
	}

	// O_EXCL used with O_CREATE, file must not exist
	file, err := os.OpenFile(ufile, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		e.GetLogger().Fatalln("Open uuid file failed of os.OpenFile(), info:", err)
	}
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		e.GetLogger().Fatalln("Read uuid file failed of os.OpenFile(), info:", err)
	}
	needCreate := true
	if len(buf) > 0 {
		_, err = uuid.Parse(string(buf))
		if err == nil {
			needCreate = false

			e.UUID = string(buf)
			e.GetLogger().Println("Fetch uuid of server Success:", e.UUID)
		}
	}

	if needCreate {
		e.UUID = uuid.NewV1().String()
		e.GetLogger().Println("Fetch uuid of server failed, Create new:", e.UUID)

		file.Write(bytes.NewBufferString(e.UUID).Bytes())
		file.Sync()
	}

	file.Close()

	return e.UUID
}

func (e *BaseInfo) GetUUID() string {
	return e.UUID
}

func (e *BaseInfo) GetQurorum() int {
	return e.Qurorum
}

func (e *BaseInfo) GetTimeout() time.Duration {
	return e.Timeout
}

func (e *BaseInfo) GetService() string {
	if e == nil {
		return ""
	}

	return e.Service
}

func (e *BaseInfo) GetLogger() *loglocal.BufferedFileLogger {
	return e.Logger
}

func (e *BaseInfo) GetNodename() string {
	return e.Service + "-" + e.LocalIP.String()
}

func (e *BaseInfo) GetHostname() string {
	host, err := os.Hostname()
	if err != nil {
		host = e.GetLocalIP().String()
	}

	return host
}

// Get Pid file path
func (e *BaseInfo) GetPidPath() string {
	return e.PidPath
}

func (e *BaseInfo) GetLocalIP() net.IP {
	return e.LocalIP
}

func (e *BaseInfo) RegisterSignalWatch() {
	defer e.Logger.Sync()

	sg := utility.NewSignalCatcher()
	stack := utility.NewSignalCallStack()
	sg.SetDefault(stack)
	sg.EnableExit()

	call := utility.NewSignalCallback(func(sig os.Signal, data interface{}) {
		defer e.Logger.Sync()
		e.Logger.Println("Receive signal: " + sig.String() + " App will terminate, bye.")
	}, nil)
	stack.Add(call)

	e.Logger.Println("Init System SignalWatcher, catch list:", strings.Join(sg.GetSignalStringList(), ", "))

	//register
	e.Sc = sg

	sg.RegisterAndServe()
}

func (e *BaseInfo) GuaranteeSingleRun() {
	defer e.GetLogger().Sync()

	pid := e.GetPidPath() + "/" + e.GetService() + ".pid"

	e.GetLogger().Println("Write and lock pid file:", pid)

	//pid create dir if needed
	err := os.MkdirAll(filepath.Dir(pid), loglocal.DEFAULT_LOGDIR_MODE)
	if err != nil {
		e.GetLogger().Fatalln("Create pid dir error, info:", err)
	}

	// O_EXCL used with O_CREATE, file must not exist
	file, err := os.OpenFile(pid, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		e.GetLogger().Fatalln("Create pid file failed of os.OpenFile(), info:", err)
	}
	procid := os.Getpid()

	//The file descriptor is valid only until f.Close is called or f is garbage collected.
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		e.GetLogger().Fatalln("syscall.Flock pid error, info:", err)
	}

	file.Write(bytes.NewBufferString(strconv.Itoa(procid)).Bytes())
	file.Sync()

	//No need to close
	//file.Close()
}
