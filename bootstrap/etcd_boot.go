package bootstrap

import (
	"os/exec"

	"zooinit/utility"
)

func BootstrapEtcd(env *envInfo) (error) {
	env.logger.Println("Starting to boot Etcd...")

	// Internal discovery service
	internalClientUrl := "http://" + env.internalHost + ":" + env.internalPort
	if env.isSelfIp {
		internalPeerUrl := "http://" + env.internalHost + ":" + env.internalPeer

		env.logger.Println("Etcd Internal PeerUrl:", internalPeerUrl)
		env.logger.Println("Etcd Internal ClientUrl:", internalClientUrl)

		// Add & can't fast wait
		intExecCmd := env.cmd + " -name " + "etcd.initial" +
		" -initial-advertise-peer-urls " + internalPeerUrl +
		" -listen-peer-urls " + internalPeerUrl +
		" -listen-client-urls " + internalClientUrl +
		" -advertise-client-urls " + internalClientUrl

		env.logger.Println("Etcd Internal ExecCmd:", intExecCmd)

		// Boot internal discovery service
		path, args, err:=utility.ParseCmdStringWithParams(intExecCmd)
		if err!=nil {
			env.logger.Fatalln("Error ParseCmdStringWithParams:", err)
		}

		cmd:=exec.Command(path, args...)
		err=cmd.Start()

		if err!=nil {
			env.logger.Fatalln("Exec Internal ExecCmd Error:", err)
		}else{
			env.logger.Println("Exec Internal OK, PID:", cmd.Process.Pid)

			// Set PID
			env.internalCmdInstance=cmd
			env.logger.Println("Internal service started.")


		}
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
	//path, args:=utility.ParseCmdStringWithParams(disExecCmd)






	return nil
}

