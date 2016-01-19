package main

import (
	"os"
	"github.com/codegangsta/cli"
	"zooinit/bootstrap"
)

func main() {
	app := cli.NewApp()
	app.Version="0.0.1"
	app.Commands = []cli.Command{
		{
			Name:      "bootstrap",
			Aliases:     []string{"boot"},
			Usage:     "Usage: "+os.Args[0]+" discovery_service_ip \nBootstrop the basic etcd based high available discovery service for low level use.",
			Action: bootstrap.Bootstrap,
		},
	}
	app.Run(os.Args)
}