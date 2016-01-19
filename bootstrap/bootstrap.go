package bootstrap

import (
	"github.com/codegangsta/cli"
	"log"
)

func Bootstrap(c *cli.Context){
	if len(c.Args())!=1 {
		log.Fatalln("bootstrap.Bootstrap: discovery_service_ip:port arg needed.")
	}
}
