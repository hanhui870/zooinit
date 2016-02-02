package cluster

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"

	"strings"
	"zooinit/cluster/etcd"
	"zooinit/config"
)

const (
	CONFIG_SECTION            = "system.cluster"
	CLUSTER_BOOTSTRAP_TIMEOUT = 5 * time.Minute
)

var (
	env *envInfo

	// Discovery service latest endpoints
	lastestEndpoints  []string
	endpointsSyncLock sync.Mutex
)

func Bootstrap(c *cli.Context) {
	fname := config.GetConfigFileName(c.String("config"))
	iniobj := config.GetConfigInstance(fname)

	if len(c.Args()) != 1 {
		fmt.Println(c.Command.Usage)
		os.Exit(1)
	}

	cluster := c.Args()[0]
	env = NewEnvInfo(iniobj, cluster)

	env.logger.Println("Logger path:", env.logPath)
	env.logger.Println("Timeout:", env.timeout.String())
	env.logger.Println("Qurorum:", env.qurorum)
	env.logger.Println("Discover method:", env.discoveryMethod)
	env.logger.Println("Discover path:", env.discoveryPath)
	env.logger.Println("env.discoveryTarget for fetch members:", env.discoveryTarget)

	// update endpoints
	UpdateLatestEndpoints()

	initializeClusterDiscoveryInfo()
}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}

// init cluster bootstrap info
func initializeClusterDiscoveryInfo() {
	kvApi, err := etcd.NewApiKeys(lastestEndpoints)
	if err != nil {
		env.logger.Fatalln("Etcd.NewApiKeys() found error:", err)
	}

	// Set qurorum size
	resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath, "", &client.SetOptions{Dir: true, TTL: CLUSTER_BOOTSTRAP_TIMEOUT, PrevExist: client.PrevNoExist})
	if err != nil {
		//ignore exist error
		if !etcd.EqualEtcdError(err, client.ErrorCodeNodeExist) {
			// check if exist need to add qurorum
			env.logger.Fatalln("Etcd.Api() set "+env.discoveryPath+" error:", err)
		} else {
			env.logger.Println("Etcd.Api() set " + env.discoveryPath + " notice: node exist, will add qurorum directly.")
		}

	} else {
		env.logger.Println("Etcd.Api() set "+env.discoveryPath+" ok:", resp)
	}

}

func UpdateLatestEndpoints() {
	memApi, err := etcd.NewApiMember(strings.Split(env.discoveryTarget, ","))
	if err != nil {
		env.logger.Fatalln("Etcd.NewApiMember() found error:", err)
	}
	tmpEndpoints, err := memApi.GetInitialClusterEndpoints()
	if err != nil {
		env.logger.Fatalln("Etcd.GetInitialClusterEndpoints() found error:", err)
	}
	env.logger.Println("Fetch discovery service latest endpoints:", tmpEndpoints)

	// lock for update
	endpointsSyncLock.Lock()
	defer endpointsSyncLock.Unlock()
	lastestEndpoints = tmpEndpoints
}
