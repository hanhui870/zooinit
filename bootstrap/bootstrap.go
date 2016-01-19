package bootstrap

import (
	"github.com/codegangsta/cli"
	"log"
	"errors"
)

var(
	configFile string
)

func Bootstrap(c *cli.Context){
	if len(c.Args())!=1 {
		log.Fatalln("bootstrap.Bootstrap: discovery_service_ip:port arg needed.")
	}
}

func GetConfigFileName() (string, error){
	if configFile=="" {
		return nil, errors.New("Bootstrap configFile path has not set.")
	}
	return configFile, nil
}
