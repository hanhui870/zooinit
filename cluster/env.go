package cluster

import (
	"bytes"
	"net"
	"os"
	"strconv"
	"syscall"
	loglocal "zooinit/log"
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
