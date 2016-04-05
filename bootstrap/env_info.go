package bootstrap

import (
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"

	"zooinit/cluster"
	"zooinit/config"
	"zooinit/utility"
)

// This basic discovery service bootstrap env info
type envInfo struct {
	cluster.BaseInfo

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

	// boot command
	cmd          string
	cmdDataDir   string
	cmdWalDir    string
	cmdSnapCount int
}

// New env from file
func NewEnvInfoFile(fname string) *envInfo {
	iniobj := config.GetConfigInstance(fname)

	return NewEnvInfo(iniobj, nil)
}

func NewEnvInfo(iniobj *ini.File, c *cli.Context) *envInfo {
	obj := new(envInfo)
	//create uuid
	obj.CreateUUID()

	sec := iniobj.Section(CONFIG_SECTION)
	obj.Service = sec.Key("service").String()
	if obj.Service == "" {
		log.Fatalln("Config of service is empty.")
	}

	// parse base info
	obj.ParseConfigFile(sec, c)

	//flush last log info
	defer obj.Logger.Sync()

	// key for process now
	var keyNow string
	obj.Logger.Println("Configure file parsed. Waiting to be boostrapped.")

	keyNow = "discovery"
	discovery := config.GetValueString(keyNow, sec, c)
	if discovery == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	if strings.Count(discovery, ":") != 2 {
		obj.Logger.Fatalln("Config of " + keyNow + " need ip:port:peer format.")
	}
	obj.discoveryHost = discovery[0:strings.Index(discovery, ":")]
	obj.discoveryPort = discovery[strings.Index(discovery, ":")+1 : strings.LastIndex(discovery, ":")]
	obj.discoveryPeer = discovery[strings.LastIndex(discovery, ":")+1:]

	keyNow = "internal"
	internal := config.GetValueString(keyNow, sec, c)
	if internal == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	if strings.Count(internal, ":") != 1 {
		obj.Logger.Fatalln("Config of " + keyNow + " need port:peer format.")
	}
	// Must be identical with discoveryHost
	obj.internalHost = obj.discoveryHost
	obj.internalPort = internal[0:strings.Index(internal, ":")]
	obj.internalPeer = internal[strings.LastIndex(internal, ":")+1:]

	keyNow = "internal.data.dir"
	path := config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.internalDataDir = path

	keyNow = "internal.wal.dir"
	path = config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.internalWalDir = path

	keyNow = "boot.cmd"
	obj.cmd = config.GetValueString(keyNow, sec, c)
	if obj.cmd == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}

	keyNow = "boot.data.dir"
	path = config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.cmdDataDir = path

	keyNow = "boot.wal.dir"
	path = config.GetValueString(keyNow, sec, c)
	if path == "" {
		obj.Logger.Fatalln("Config of " + keyNow + " is empty.")
	}
	obj.cmdWalDir = path

	keyNow = "boot.snap.count"
	snapCount, err := config.GetValueFloat64(keyNow, sec, c)
	if err != nil {
		obj.Logger.Fatalln("Config of etcd "+keyNow+" is error:", err)
	}
	if snapCount > 100000 || snapCount < 100 {
		obj.Logger.Fatalln("Config of etcd " + keyNow + " must between 100-100000")
	} else {
		obj.cmdSnapCount = int(snapCount)
	}

	// Init Extra runtime info
	if utility.HasIpAddress(obj.internalHost) {
		obj.isSelfIp = true
		obj.LocalIP = net.ParseIP(obj.internalHost)
	} else {
		obj.isSelfIp = false

		localip, err := utility.GetLocalIPWithIntranet(obj.internalHost)
		if err != nil {
			obj.Logger.Fatalln("utility.GetLocalIPWithIntranet Please check configuration of discovery is correct.")
		}
		obj.LocalIP = localip
	}

	return obj
}

// Fetch bootstrap command
func (e *envInfo) GetCmd() string {
	return e.cmd
}

func (e *envInfo) GetDiscoveryHost() string {
	return e.discoveryHost
}

func (e *envInfo) GetDiscoveryPort() string {
	return e.discoveryPort
}

func (e *envInfo) GetInternalClientUrl() string {
	return "http://" + e.internalHost + ":" + e.internalPort
}

func (e *envInfo) GetInternalPeerUrl() string {
	return "http://" + e.internalHost + ":" + e.internalPeer
}

func (e *envInfo) GetClientUrl() string {
	return "http://" + env.LocalIP.String() + ":" + env.discoveryPort
}

func (e *envInfo) GetPeerUrl() string {
	return "http://" + env.LocalIP.String() + ":" + env.discoveryPeer
}
