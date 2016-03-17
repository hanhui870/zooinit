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

	// Common flags
	cfgFlag := &cli.StringFlag{
		Name:  "config, f",
		Value: "config/config.ini",
		Usage: "Configuration file of zooinit.",
	}

	qurorum := &cli.IntFlag{
		Name:  "qurorum, q",
		Value: 0,
		Usage: "Predefined qurorum of cluster members waiting to be booted.",
	}

	healthCheck := &cli.IntFlag{
		Name:  "health.check.interval, n",
		Value: 0,
		Usage: "Health check interval, default 2 sec, same to zookeeper ticktime.",
	}

	timeout := &cli.IntFlag{
		Name:  "timeout, t",
		Value: 0,
		Usage: "Cluster bootstrap timeout, sec unit.",
	}

	logChannel := &cli.StringFlag{
		Name:  "log.channel",
		Value: "",
		Usage: "Configuration of runtime log channel: file, write to file; stdout, write to stdout; multi, write both.",
	}

	logPath := &cli.StringFlag{
		Name:  "log.path, log",
		Value: "",
		Usage: "Configuration of runtime log path.",
	}

	// Used for cluster
	backendFlag := &cli.StringFlag{
		Name:  "backend, b",
		Value: "",
		Usage: "Backend name of cluster, eg: consul, etcd, zookeeper...",
	}

	ipHint := &cli.StringFlag{
		Name:  "ip.hint",
		Value: "",
		Usage: "Ip hint use to found which ip for boot bind, will automatically find intranet ip.",
	}

	discoverMethod := &cli.StringFlag{
		Name:  "discover.method",
		Value: "",
		Usage: "Available: address, may single point failure; dnssrv, this could be a second choise, with dnssrv update api.",
	}

	discoverTarget := &cli.StringFlag{
		Name:  "discover.target",
		Value: "",
		Usage: "Discovery target value to discover other members.",
	}

	discoverPathPrefix := &cli.StringFlag{
		Name:  "discover.path.prefix",
		Value: "",
		Usage: "Ip hint use to found which ip for boot bind, will automatically find intranet ip.",
	}

	// Used for bootstrap etcd
	discovery := &cli.StringFlag{
		Name:  "discovery, d",
		Value: "",
		Usage: "Bootstrap etcd cluster service for boot other cluster service, also iphint to find locale ip. fomat: ip:port:peer",
	}

	internal := &cli.StringFlag{
		Name:  "internal",
		Value: "",
		Usage: "The same IP with discovery. Ports distinct from discovery in the same machine.",
	}

	internalData := &cli.StringFlag{
		Name:  "internal.data.dir",
		Value: "",
		Usage: "Used for internal bootstrap for system, Only one member. etcd.data.dir.",
	}

	internalWal := &cli.StringFlag{
		Name:  "internal.wal.dir",
		Value: "",
		Usage: "Used for internal bootstrap for system, Only one member. etcd.wal.dir.",
	}

	bootcmd := &cli.StringFlag{
		Name:  "boot.cmd",
		Value: "",
		Usage: "The same IP with discovery. Ports distinct from discovery in the same machine.",
	}

	bootData := &cli.StringFlag{
		Name:  "boot.data.dir",
		Value: "",
		Usage: "Used for bootstrap cluster, Only one member. boot.data.dir.",
	}

	bootWal := &cli.StringFlag{
		Name:  "boot.wal.dir",
		Value: "",
		Usage: "Used for bootstrap cluster, Only one member. boot.wal.dir.",
	}

	bootSnap := &cli.StringFlag{
		Name:  "boot.snap.count",
		Value: "",
		Usage: "Used for bootstrap cluster, Only one member, Etcd config --snapshot-count '10000'. boot.snap.count.",
	}

	app.Commands = []cli.Command{
		{
			Name:    "bootstrap",
			Aliases: []string{"boot"},
			Usage:   "Usage: " + os.Args[0] + " bootstrap -f config.ini \nBootstrop the basic etcd based high available discovery service for low level use.",
			Action:  bootstrap.Bootstrap,
			Flags: []cli.Flag{
				discovery, internal, internalData, internalWal,
				bootcmd, bootData, bootWal, bootSnap,
				cfgFlag, qurorum, healthCheck, timeout, logChannel, logPath,
			},
		},
		{
			Name:   "cluster",
			Usage:  "Usage: " + os.Args[0] + " cluster -f config.ini -b backend service \nBootstrop the cluster configured in the configuration file.",
			Action: cluster.Bootstrap,
			Flags: []cli.Flag{
				backendFlag, ipHint, discoverMethod, discoverTarget, discoverPathPrefix,
				cfgFlag, qurorum, healthCheck, timeout, logChannel, logPath,
			},
		},
	}
	app.Run(os.Args)
}
