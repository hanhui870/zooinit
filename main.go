package main

import (
	"os"
	"github.com/codegangsta/cli"
	"zooinit/bootstrap"
)

func main() {
	app := cli.NewApp()
	app.Version="0.0.1"

	cfgFlag:=&cli.StringFlag{
		Name:  "config, f",
		Value: "config/config.ini",
		Usage: "Configuration file of zooini.",
	}

	app.Commands = []cli.Command{
		{
			Name:      "bootstrap",
			Aliases:     []string{"boot"},
			Usage:     "Usage: "+os.Args[0]+" bootstrap -f config.ini \nBootstrop the basic etcd based high available discovery service for low level use.",
			Action: bootstrap.Bootstrap,
			Flags: []cli.Flag{
				cfgFlag,
			},
		},
		{
			Name:      "cluster",
			Usage:     "Usage: "+os.Args[0]+" cluster -f config.ini clustername \nBootstrop the cluster configured in the configuration file.",
			Action: bootstrap.Bootstrap,
			Flags: []cli.Flag{
				cfgFlag,
			},
		},
	}
	app.Run(os.Args)
}