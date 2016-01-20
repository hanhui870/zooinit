package bootstrap

import (
	"log"
	"os"

	"github.com/go-ini/ini"
	"github.com/codegangsta/cli"

	"path/filepath"
	"zooinit/config"
)

const (
	CONFIG_SECTION = "system.boostrap"
)

var (
	env *envInfo
)

func Bootstrap(c *cli.Context) {
	fname := GetConfigFileName(c.String("config"))
	iniobj := GetConfigInstance(fname)

	env=NewEnvInfo(iniobj)
}

// Fetch bootstrap env instance
func GetEnvInfo()(*envInfo){
	return env
}

func GetConfigFileName(configfile string) (string) {
	if configfile == "" {
		log.Fatalln("Bootstrap configFile path has not set.")
	}

	fname, err := filepath.Abs(configfile)
	if err != nil {
		log.Fatalln("Fetch Abs config file error:", err)
	}
	log.Println("Use configuration file:", fname)

	return fname
}

func GetConfigInstance(fname string) (*ini.File) {
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		log.Fatalln("Configuration file:", err)
	}

	iniInstance := config.Ini(fname)
	return iniInstance
}


