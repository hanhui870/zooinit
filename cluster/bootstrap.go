package cluster

import (
	"fmt"
	"os"
	"time"

	"zooinit/config"

	"github.com/codegangsta/cli"
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

	println("hello world.")
}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}
