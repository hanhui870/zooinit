package bootstrap

import (
	"os/exec"
)

func BootstrapEtcd(env *envInfo) (error) {
	env.logger.Println("Starting to boot Etcd...")

	// Internal discovery service
	internalClientUrl := "http://" + env.internalHost + ":" + env.internalPort
	if env.isSelfIp {
		internalPeerUrl := "http://" + env.internalHost + ":" + env.internalPeer

		env.logger.Println("Etcd Internal PeerUrl:", internalPeerUrl)
		env.logger.Println("Etcd Internal ClientUrl:", internalClientUrl)

		intExecCmd := env.cmd + " -name " + "etcd.initial" +
		" -initial-advertise-peer-urls " + internalPeerUrl +
		" -listen-peer-urls " + internalPeerUrl +
		" -listen-client-urls http://127.0.0.1:2379," + internalClientUrl +
		" -advertise-client-urls " + internalClientUrl

		env.logger.Println("Etcd Internal ExecCmd:", intExecCmd)
	}

	// Cluster member startup info
	discoveryPeerUrl := "http://" + env.localIP.String() + ":" + env.discoveryPeer
	discoveryClientUrl := "http://" + env.localIP.String() + ":" + env.discoveryPort
	env.logger.Println("Etcd Discovery PeerUrl:", discoveryPeerUrl)
	env.logger.Println("Etcd Discovery ClientUrl:", discoveryClientUrl)

	disExecCmd := env.cmd + " -name " + "etcd.bootstrap." +
	" -initial-advertise-peer-urls " + discoveryPeerUrl +
	" -listen-peer-urls " + discoveryPeerUrl +
	" -listen-client-urls http://127.0.0.1:2379," + discoveryClientUrl +
	" -advertise-client-urls " + discoveryClientUrl

	env.logger.Println("Etcd Discovery ExecCmd:", disExecCmd)


	// Boot internal discovery service
	exec.Cmd{}




	return nil
}

