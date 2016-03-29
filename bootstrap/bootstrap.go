package bootstrap

import (
	"github.com/codegangsta/cli"

	"zooinit/cluster"
	"zooinit/config"
)

const (
	CONFIG_SECTION = "system.boostrap"
)

var (
	env *envInfo
)

func Bootstrap(c *cli.Context) {
	fname := config.GetConfigFileName(c.String("config"))
	iniobj := config.GetConfigInstance(fname)

	env = NewEnvInfo(iniobj, c)

	cluster.GuaranteeSingleRun(env)

	//register signal watcher
	env.registerSignalWatch()

	BootstrapEtcd(env)
}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}
