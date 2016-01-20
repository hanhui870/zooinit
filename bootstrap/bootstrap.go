package bootstrap

import (
	"github.com/codegangsta/cli"
	"log"
	"path/filepath"
	"os"
)

func Bootstrap(c *cli.Context) {
	fname := GetConfigFileName(c.String("config"))



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

	_, err = os.Stat(fname)
	if os.IsNotExist(err) {
		log.Fatalln("Configuration file:", err)
	}

	return fname
}


