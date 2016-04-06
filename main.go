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

	pidPath := &cli.StringFlag{
		Name:  "pid.path, pid",
		Value: "",
		Usage: "Configuration of runtime log path.",
	}

	// args used to find local ip.
	ipLocal := &cli.StringFlag{
		Name:  "ip.local",
		Value: "",
		Usage: "Ip local for boot ip, overwrite all other setting.",
	}
	ipMethod := &cli.StringFlag{
		Name:  "ip.method",
		Value: "",
		Usage: "Ip method use to find ip for boot bind, available:default, netmask, interface. See config.ini for detail.",
	}
	ipHint := &cli.StringFlag{
		Name:  "ip.hint",
		Value: "",
		Usage: "Ip hint for find ip, with real the same netmask. ip.method: default.",
	}
	ipNetmask := &cli.StringFlag{
		Name:  "ip.netmask",
		Value: "",
		Usage: "Ip netmask for find ip, with -ip.netmask=255.0.0.0. ip.method: netmask.",
	}
	ipInterface := &cli.StringFlag{
		Name:  "ip.interface",
		Value: "",
		Usage: "Ip interface for find ip, with -ip.interface=eth0. wip.method: interface.",
	}

	// Used for cluster
	backendFlag := &cli.StringFlag{
		Name:  "backend, b",
		Value: "",
		Usage: "Backend name of cluster, eg: consul, etcd, zookeeper...",
	}

	discoverMethod := &cli.StringFlag{
		Name:  "discovery.method",
		Value: "",
		Usage: "Available: address, may single point failure; dnssrv, this could be a second choise, with dnssrv update api.",
	}

	discoverTarget := &cli.StringFlag{
		Name:  "discovery.target",
		Value: "",
		Usage: "Discovery target value to discover other members.",
	}

	discoverPathPrefix := &cli.StringFlag{
		Name:  "discovery.path.prefix",
		Value: "",
		Usage: "Discovery service path prefix registered in etcd boot cluster.",
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
				cfgFlag, qurorum, healthCheck, timeout, logChannel, logPath, pidPath, ipLocal, ipMethod, ipHint, ipNetmask, ipInterface,
			},
		},
		{
			Name:   "cluster",
			Usage:  "Usage: " + os.Args[0] + " cluster -f config.ini -b backend service \nBootstrop the cluster configured in the configuration file.",
			Action: cluster.Bootstrap,
			Flags: []cli.Flag{
				backendFlag, discoverMethod, discoverTarget, discoverPathPrefix,
				cfgFlag, qurorum, healthCheck, timeout, logChannel, logPath, pidPath, ipLocal, ipMethod, ipHint, ipNetmask, ipInterface,
			},
		},
	}
	app.Run(os.Args)
}
