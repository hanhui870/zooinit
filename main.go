package main

import (
	"github.com/codegangsta/cli"
	"os"
	"zooinit/bootstrap"
	"zooinit/cluster"
)

func main() {
	app := cli.NewApp()
	app.Author = "bruce"
	app.Email = "bruce@haimi.com"
	app.Copyright = "haimi.com All rights reseverd."
	app.Name = "Zooinit"
	app.Usage = "An high available service for Zookeeper/Consul/Hadoop alike clusters bootstraping and being watched."
	app.Version = "0.0.9"

	cfgFlag := &cli.StringFlag{
		Name:  "config, f",
		Value: "config/config.ini",
		Usage: "Configuration file of zooini.",
	}

	backendFlag := &cli.StringFlag{
		Name:  "backend, b",
		Value: "",
		Usage: "Backend name of cluster, eg: consul, etcd, zookeeper...",
	}

	app.Commands = []cli.Command{
		{
			Name:    "bootstrap",
			Aliases: []string{"boot"},
			Usage:   "Usage: " + os.Args[0] + " bootstrap -f config.ini \nBootstrop the basic etcd based high available discovery service for low level use.",
			Action:  bootstrap.Bootstrap,
			Flags: []cli.Flag{
				cfgFlag,
			},
		},
		{
			Name:   "cluster",
			Usage:  "Usage: " + os.Args[0] + " cluster -f config.ini -b backend service \nBootstrop the cluster configured in the configuration file.",
			Action: cluster.Bootstrap,
			Flags: []cli.Flag{
				cfgFlag, backendFlag,
			},
		},
	}
	app.Run(os.Args)
}
