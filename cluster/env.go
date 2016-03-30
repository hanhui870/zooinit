package cluster

import (
	"bytes"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

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
}

func GuaranteeSingleRun(env Env) {
	defer env.GetLogger().Sync()

	pid := env.GetPidPath() + "/" + env.GetService() + ".pid"

	env.GetLogger().Println("Write and lock pid file:", pid)

	// O_EXCL used with O_CREATE, file must not exist
	file, err := os.OpenFile(pid, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		env.GetLogger().Fatalln("Create pid file failed of os.OpenFile(), info:", err)
	}
	procid := os.Getpid()

	//The file descriptor is valid only until f.Close is called or f is garbage collected.
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		env.GetLogger().Fatalln("syscall.Flock pid error, info:", err)
	}

	file.Write(bytes.NewBufferString(strconv.Itoa(procid)).Bytes())
	file.Sync()

	//No need to close
	//file.Close()
}

type BaseInfo struct {
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
