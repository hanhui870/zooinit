package bootstrap

import (
	"github.com/codegangsta/cli"

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

	//register signal watcher
	env.registerSignalWatch()

	BootstrapEtcd(env)
}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}
