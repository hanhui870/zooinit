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
	Logger() *loglocal.BufferedFileLogger

	// localIP for boot
	LocalIP() net.IP

	// service name, also use for log
	Service() string

	// Get Pid file path
	GetPidPath() string
}

func GuaranteeSingleRun(env Env) {
	defer env.Logger().Sync()

	pid := env.GetPidPath() + "/" + env.Service() + ".pid"

	env.Logger().Println("Write and lock pid file:", pid)

	// O_EXCL used with O_CREATE, file must not exist
	file, err := os.OpenFile(pid, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		env.Logger().Fatalln("Create pid file failed of os.OpenFile(), info:", err)
	}
	procid := os.Getpid()

	//The file descriptor is valid only until f.Close is called or f is garbage collected.
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		env.Logger().Fatalln("syscall.Flock pid error, info:", err)
	}

	file.Write(bytes.NewBufferString(strconv.Itoa(procid)).Bytes())
	file.Sync()

	//No need to close
	//file.Close()
}
