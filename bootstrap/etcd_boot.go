package bootstrap

import (
	"os/exec"
	"time"
	"strconv"
	"net/http"

	"github.com/coreos/etcd/client"

	"zooinit/utility"
	"zooinit/cluster/etcd"
	"zooinit/log"
)

const (
// INTERNAL discovery findpath
	INTERNAL_FINDPATH = "/boot/initial"
	CLUSTER_BOOTSTRAP_TIMEOUT = 5 * time.Minute
)

func BootstrapEtcd(env *envInfo) (error) {
	env.logger.Println("Starting to boot Etcd...")

	// Internal discovery service
	internalClientUrl := "http://" + env.internalHost + ":" + env.internalPort
	// Api to internal service
	api, err := etcd.NewApi([]string{internalClientUrl})
	if err != nil {
		env.logger.Fatal("Etcd NewApi error:", err)
	}

	if env.isSelfIp {
		internalPeerUrl := "http://" + env.internalHost + ":" + env.internalPeer

		env.logger.Println("Etcd Internal PeerUrl:", internalPeerUrl)
		env.logger.Println("Etcd Internal ClientUrl:", internalClientUrl)

		// Add & can't fast wait
		// data-dir can't be same with discovery service.
		intName := "etcd.initial"
		intExecCmd := "etcd --data-dir /tmp/internal/etcd/data -wal-dir /tmp/internal/etcd/wal -name " + intName +
		" -initial-advertise-peer-urls " + internalPeerUrl +
		" -listen-peer-urls " + internalPeerUrl +
		" -listen-client-urls " + internalClientUrl +
		" -advertise-client-urls " + internalClientUrl +
		" -initial-cluster " + intName + "=" + internalPeerUrl

		env.logger.Println("Etcd Internal ExecCmd:", intExecCmd)

		// Boot internal discovery service
		path, args, err := utility.ParseCmdStringWithParams(intExecCmd)
		if err != nil {
			env.logger.Fatalln("Error ParseCmdStringWithParams internal service:", err)
		}

		cmd := exec.Command(path, args...)
		loggerIOAdapter := log.NewLoggerIOAdapter(env.logger)
		loggerIOAdapter.SetPrefix("Etcd internal server: ")
		cmd.Stdout = loggerIOAdapter
		cmd.Stderr = loggerIOAdapter
		err = cmd.Start()


		if err != nil {
			env.logger.Fatalln("Exec Internal ExecCmd Error:", err)
		}else {
			env.logger.Println("Exec Internal OK, PID:", cmd.Process.Pid)

			// Release process after cluster up.
			defer cmd.Process.Kill()

			// Set PID
			env.internalCmdInstance = cmd
			env.logger.Println("Internal service started.")

			// Important!!!
			tts := 3 * time.Second
			env.logger.Println("Etcd internal Sleep ", tts.String(), " for startup...")
			time.Sleep(tts)

			resp, err := http.Get(internalClientUrl + "/v2/stats/self")
			if err != nil {
				env.logger.Fatal("Error fetch stats self:", err)
			}
			env.logger.Println("Etcd internal Stat self:", resp)

			_, err = api.Conn().Delete(etcd.Context(), INTERNAL_FINDPATH, &client.DeleteOptions{Dir:true, Recursive:true})
			if err != nil {
				//type safe cast
				err, ok := err.(client.Error)
				if ok && err.Code != client.ErrorCodeKeyNotFound {
					env.logger.Fatal("Delete ", INTERNAL_FINDPATH, " error:", err)
				}
			}

			_, err = api.Conn().Set(etcd.Context(), INTERNAL_FINDPATH, "", &client.SetOptions{TTL:CLUSTER_BOOTSTRAP_TIMEOUT, Dir:true})
			if err != nil {
				env.logger.Fatal("Set TTL for ", INTERNAL_FINDPATH, " error:", err)
			}

			env.logger.Println("Set Qurorum ", INTERNAL_FINDPATH + "/_config/size to", env.qurorum)
			_, err = api.Conn().Set(etcd.Context(), INTERNAL_FINDPATH + "/_config/size", strconv.Itoa(env.qurorum), nil)
			if err != nil {
				env.logger.Fatal("Set Qurorum ", INTERNAL_FINDPATH + "/_config/size error:", err)
			}
		}
	}

	// Cluster member startup info
	discoveryPeerUrl := "http://" + env.localIP.String() + ":" + env.discoveryPeer
	discoveryClientUrl := "http://" + env.localIP.String() + ":" + env.discoveryPort
	env.logger.Println("Etcd Discovery PeerUrl:", discoveryPeerUrl)
	env.logger.Println("Etcd Discovery ClientUrl:", discoveryClientUrl)

	disExecCmd := env.cmd + " -name " + "etcd.bootstrap." + env.localIP.String() +
	" -initial-advertise-peer-urls " + discoveryPeerUrl +
	" -listen-peer-urls " + discoveryPeerUrl +
	" -listen-client-urls http://127.0.0.1:2379," + discoveryClientUrl +
	" -advertise-client-urls " + discoveryClientUrl +
	" -discovery " + internalClientUrl + "/v2/keys" + INTERNAL_FINDPATH

	env.logger.Println("Etcd Discovery ExecCmd:", disExecCmd)


	// Boot internal discovery service
	// Need to rm -rf /tmp/etcd/ because dir may be used before
	path, args, err := utility.ParseCmdStringWithParams(disExecCmd)
	if err != nil {
		env.logger.Fatalln("Error ParseCmdStringWithParams cluster bootstrap:", err)
	}

	cmd := exec.Command(path, args...)
	loggerIOAdapter := log.NewLoggerIOAdapter(env.logger)
	loggerIOAdapter.SetPrefix("Etcd discovery member: ")
	cmd.Stdout = loggerIOAdapter
	cmd.Stderr = loggerIOAdapter
	err = cmd.Start()

	if err != nil {
		env.logger.Fatalln("Exec Discovery ExecCmd Error:", err)
	}else {
		env.logger.Println("Exec Discovery Etcd member OK, PID:", cmd.Process.Pid)
		env.logger.Println("Etcd member service ", discoveryClientUrl, " started,  waiting to be bootrapped.")
	}

	// If stoped, process's output can't trace no longer
	time.Sleep(CLUSTER_BOOTSTRAP_TIMEOUT)
	return nil
}

