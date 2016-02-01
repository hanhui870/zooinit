package cluster

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"

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
	memApi, err := etcd.NewApiMember(strings.Split(env.discoveryTarget, ","))
	if err != nil {
		env.logger.Fatalln("Etcd.NewApiMember() found error:", err)
	}
	config, err := memApi.GetInitialClusterEndpoints()
	env.logger.Println("Discovery service find latest endpoints:", config)
	kvApi, err := etcd.NewApiKeys(config)

	kvApi.Conn()

}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}
