package bootstrap

import (
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/client"

	"zooinit/cluster"
	"zooinit/cluster/etcd"
	"zooinit/log"
	"zooinit/utility"
)

const (
	// INTERNAL discovery findpath
	INTERNAL_FINDPATH         = "/zooinit/boot"
	CLUSTER_BOOTSTRAP_TIMEOUT = 5 * time.Minute

	DEFAULT_BOOTSTRAP_DISCOVERY_PATH = "/zooinit/discovery/cluster/" + cluster.BOOTSTRAP_SERVICE_NAME
)

var (
	// Cluster run cmd
	clusterCmd *exec.Cmd
	// Internal cmd run
	internalCmd *exec.Cmd

	// Restart cluster member channel
	restartMemberChannel chan int
	execCheckFailedTimes int

	// Is terminate the app
	exitApp atomic.Value

	// WaitGroup for goroutine to complete
	cmdWaitGroup sync.WaitGroup

	// Discovery service latest endpoints
	lastestEndpoints  []string
	endpointsSyncLock sync.Mutex
)

func init() {
	// init channel
	restartMemberChannel = make(chan int)
}

func BootstrapEtcd(env *envInfo) {
	// flush last log info
	defer env.logger.Sync()
	env.logger.Println("Starting to boot Etcd...")

	env.logger.Println("Logger channel:", env.logChannel)
	if env.logChannel != log.LOG_STDOUT {
		env.logger.Println("Logger path:", env.logPath)
	}
	env.logger.Println("Timeout:", env.timeout.String())
	env.logger.Println("Qurorum:", env.qurorum)
	env.logger.Println("Discover internal:", env.GetInternalClientUrl())
	env.logger.Println("Health check interval:", env.healthCheckInterval)

	// Boot Internal etcd
	if env.isSelfIp {
		bootUpInternalEtcd()
	}

	// Boot local cluster member
	bootstrapLocalClusterMember()

	checkDiscoveryClusterIsUp()

	// Up and fetch latest endpoints.
	UpdateLatestEndpoints()

	//must before watchDogRunning
	go clusterMemberRestartRoutine()

	// watch and check cluster health [watchdog], block until server receive term signal
	// check cluster bootstraped and register memberself
	// If stoped, process's output can't trace no longer
	watchDogRunning()

	// final wait.
	cmdWaitGroup.Wait()

	env.logger.Println("App runtime reaches end.")
}

func bootUpInternalEtcd() {
	// flush last log info
	defer env.logger.Sync()

	// Api to internal service
	api, err := etcd.NewApiKeys([]string{env.GetInternalClientUrl()})
	if err != nil {
		env.logger.Fatal("Etcd NewApi error:", err)
	}

	env.logger.Println("Etcd Internal PeerUrl:", env.GetInternalPeerUrl())
	env.logger.Println("Etcd Internal ClientUrl:", env.GetInternalClientUrl())

	// Add & can't fast wait
	// data-dir can't be same with discovery service.
	intName := "etcd.initial"
	intExecCmd := env.cmd + " --data-dir " + env.internalDataDir + " -wal-dir " + env.internalWalDir + " -name " + intName +
		" -initial-advertise-peer-urls " + env.GetInternalPeerUrl() +
		" -listen-peer-urls " + env.GetInternalPeerUrl() +
		" -listen-client-urls " + env.GetInternalClientUrl() +
		" -advertise-client-urls " + env.GetInternalClientUrl() +
		" -initial-cluster " + intName + "=" + env.GetInternalPeerUrl()

	env.logger.Println("Etcd Internal ExecCmd:", intExecCmd)

	// Boot internal discovery service
	path, args, err := utility.ParseCmdStringWithParams(intExecCmd)
	if err != nil {
		env.logger.Fatalln("Error ParseCmdStringWithParams internal service:", err)
	}

	internalCmd = exec.Command(path, args...)
	loggerIOAdapter := log.NewLoggerIOAdapter(env.logger)
	loggerIOAdapter.SetPrefix("Etcd internal server: ")
	internalCmd.Stdout = loggerIOAdapter
	internalCmd.Stderr = loggerIOAdapter
	err = internalCmd.Start()

	// internal cmd wait
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		err = internalCmd.Wait()
		if err != nil {
			env.logger.Println("internalCmd.Wait() finished with error found:", err)
		} else {
			env.logger.Println("internalCmd.Wait() finished without error.")
		}
	}()

	if err != nil {
		env.logger.Fatalln("Exec Internal ExecCmd Error:", err)
	} else {
		env.logger.Println("Exec Internal OK, PID:", internalCmd.Process.Pid)

		// Set PID
		env.internalCmdInstance = internalCmd
		env.logger.Println("Internal service started.")

		// Important!!! check upstarted
		env.logger.Println("Etcd LoopTimeoutRequest for check internal's startup...")

		internalCheckout := 3 * time.Second
		isHealth, err := LoopTimeoutRequest(internalCheckout, env, func() (bool, error) {
			return etcd.CheckHealth(env.GetInternalClientUrl())
		})
		if err != nil {
			env.logger.Fatal("Error check internal error: ", err)
		} else if isHealth != true {
			env.logger.Fatal("Error check internal server health: ", isHealth)
			env.logger.Fatal("Cluster bootstrap faild: failed to bootstrap in ", internalCheckout.String())
		}

		resp, err := http.Get(env.GetInternalClientUrl() + "/v2/stats/self")
		if err != nil {
			env.logger.Fatal("Error fetch stats self: ", err)
		}
		env.logger.Println("Etcd internal Stat self: ", resp)

		_, err = api.Conn().Delete(etcd.Context(), INTERNAL_FINDPATH, &client.DeleteOptions{Dir: true, Recursive: true})
		if err != nil {
			//type safe cast
			err, ok := err.(client.Error)
			if ok && err.Code != client.ErrorCodeKeyNotFound {
				env.logger.Fatal("Delete ", INTERNAL_FINDPATH, " error:", err)
			}
		}

		env.logger.Println("Set Cluster bootstrap timeout:", env.timeout.String())
		_, err = api.Conn().Set(etcd.Context(), INTERNAL_FINDPATH, "", &client.SetOptions{TTL: env.timeout, Dir: true})
		if err != nil {
			env.logger.Fatal("Set TTL for ", INTERNAL_FINDPATH, " error:", err)
		}

		env.logger.Println("Set Qurorum ", INTERNAL_FINDPATH+"/_config/size to ", env.qurorum)
		_, err = api.Conn().Set(etcd.Context(), INTERNAL_FINDPATH+"/_config/size", strconv.Itoa(env.qurorum), nil)
		if err != nil {
			env.logger.Fatal("Set Qurorum ", INTERNAL_FINDPATH+"/_config/size error: ", err)
		}
	}
}

func bootstrapLocalClusterMember() {
	//flush last log info
	defer env.logger.Sync()

	// Cluster member startup info
	env.logger.Println("Etcd Discovery PeerUrl:", env.GetPeerUrl())
	env.logger.Println("Etcd Discovery ClientUrl:", env.GetClientUrl())

	// Etcd cluster can restart, because etcd restart don't need discovery service
	disExecCmd := env.cmd + " --data-dir " + env.cmdDataDir + " -wal-dir " + env.cmdWalDir +
		" -snapshot-count " + strconv.Itoa(env.cmdSnapCount) +
		" -name " + "etcd.bootstrap." + env.localIP.String() +
		" -initial-advertise-peer-urls " + env.GetPeerUrl() +
		" -listen-peer-urls " + env.GetPeerUrl() +
		" -listen-client-urls http://127.0.0.1:2379," + env.GetClientUrl() +
		" -advertise-client-urls " + env.GetClientUrl() +
		" -discovery " + env.GetInternalClientUrl() + "/v2/keys" + INTERNAL_FINDPATH

	env.logger.Println("Etcd Discovery ExecCmd: ", disExecCmd)

	// Boot internal discovery service
	// Need to rm -rf /tmp/etcd/ because dir may be used before
	path, args, err := utility.ParseCmdStringWithParams(disExecCmd)
	if err != nil {
		env.logger.Fatalln("Error ParseCmdStringWithParams cluster bootstrap: ", err)
	}

	clusterCmd = exec.Command(path, args...)
	loggerIOAdapter := log.NewLoggerIOAdapter(env.logger)
	loggerIOAdapter.SetPrefix("Etcd discovery member: ")
	clusterCmd.Stdout = loggerIOAdapter
	clusterCmd.Stderr = loggerIOAdapter

	err = clusterCmd.Start()
	if err != nil {
		env.logger.Fatalln("Exec Discovery ExecCmd Error: ", err)
	} else {
		env.logger.Println("Exec Discovery Etcd member OK, PID: ", clusterCmd.Process.Pid)
		env.logger.Println("Etcd member service ", env.GetClientUrl(), " started,  waiting to be bootrapped.")
	}

	// here no block now, block in health check
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		err = clusterCmd.Wait()
		if err != nil {
			env.logger.Println("callCmd.Wait() finished with error found:", err)
		} else {
			env.logger.Println("callCmd.Wait() finished without error.")
		}

		if isExit, ok := exitApp.Load().(bool); !ok || !isExit {
			env.logger.Println("BootstrapLocalClusterMember do not detect exitApp cmd, will restart...")
			restartMemberChannel <- cluster.MEMBER_RESTART_CMDWAIT
		}
	}()
}

func checkDiscoveryClusterIsUp() {
	//flush last log info
	defer env.logger.Sync()

	// Important!!! check upstarted
	env.logger.Println("Etcd LoopTimeoutRequest for check discovery cluster's startup...")
	isHealth, err := LoopTimeoutRequest(env.timeout, env, func() (bool, error) {
		return etcd.CheckHealth(env.GetClientUrl())
	})
	if err != nil {
		env.logger.Fatal("Error check discovery error: ", err)
	} else if isHealth != true {
		env.logger.Fatal("Error check discovery server health: ", isHealth)
		env.logger.Fatal("Cluster bootstrap faild: failed to bootstrap in ", env.timeout.String())
	}

	// Close internal service
	env.logger.Println("Cluster etcd service is booted. Internal service is going to be terminated.")
	if internalCmd != nil && internalCmd.Process != nil {
		internalCmd.Process.Kill()
	}
}

//request until sucess
func LoopTimeoutRequest(timeout time.Duration, env *envInfo, routine func() (result bool, err error)) (result bool, err error) {
	var charlist []byte

	//flush last log info
	defer env.logger.Sync()

	result = false
	start := time.Now()
	for {
		result, err = routine()
		if !result || err != nil {
			charlist = append(charlist, byte('.'))
			// sleep 100ms
			end := time.Now()
			// not time outed
			if end.Sub(start) < timeout {
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		} else {
			break
		}
	}

	env.logger.Println("Fetched data LoopTimeoutRequest for loop:", string(charlist))

	return result, err
}

func watchDogRunning() {
	//flush last log info
	defer env.logger.Sync()

	for {
		if isExit, ok := exitApp.Load().(bool); ok && isExit {
			env.logger.Println("Receive exitApp signal, break watchDogRunning loop.")
			break
		}

		execHealthChechRunning()
		//do not need break, because loop is maitained by zooinit

		// sleep interval time
		time.Sleep(env.healthCheckInterval)
	}
}

// Check cluster health after cluster is up.
func execHealthChechRunning() (result bool) {
	//flush last log info
	defer env.logger.Sync()

	var cm *cluster.ClusterMember

	isHealthy, err := etcd.CheckHealth(env.GetClientUrl())
	// ttl 1min, update 1/s
	if err == nil && isHealthy {
		// when the health check call normal return, break the infinite loop
		result = true
		// reset to 0
		execCheckFailedTimes = 0
	} else {
		result = false
		// trigger restart related
		execCheckFailedTimes++

		if err != nil {
			env.logger.Println("Found error while etcd.CheckHealth("+env.GetClientUrl()+"):", err)
		}

		if execCheckFailedTimes >= cluster.MEMBER_MAX_FAILED_TIMES {
			env.logger.Println("Cluster member is NOT healthy, will trigger Restart. Failed times:", execCheckFailedTimes, ", MEMBER_MAX_FAILED_TIMES:", cluster.MEMBER_MAX_FAILED_TIMES)
			execCheckFailedTimes = 0
			restartMemberChannel <- cluster.MEMBER_RESTART_HEALTHCHECK
		} else {
			env.logger.Println("Cluster member is NOT healthy, Failed times:", execCheckFailedTimes)
		}
	}

	kvApi := getClientKeysApi()
	cm = cluster.NewClusterMember(env.GetNodename(), env.localIP.String(), result, execCheckFailedTimes)
	pathNode := DEFAULT_BOOTSTRAP_DISCOVERY_PATH + cluster.CLUSTER_MEMBER_DIR + "/" + env.GetNodename()
	resp, err := kvApi.Conn().Set(etcd.Context(), pathNode, cm.ToJson(), &client.SetOptions{Dir: false, TTL: cluster.CLUSTER_MEMBER_NODE_TTL})
	if err != nil {
		env.logger.Fatalln("Etcd.Api() update "+pathNode+" State error:", err)
	} else {
		env.logger.Println("Etcd.Api() update "+pathNode+" ok", "Resp:", resp)
	}

	return result
}

func getClientKeysApi() *etcd.ApiKeys {
	//flush last log info
	defer env.logger.Sync()

	kvApi, err := etcd.NewApiKeys(lastestEndpoints)
	if err != nil {
		env.logger.Fatalln("Etcd.NewApiKeys() found error:", err)
	}

	return kvApi
}

func getClientMembersApi(endpoints []string) *etcd.ApiMembers {
	//flush last log info
	defer env.logger.Sync()

	memApi, err := etcd.NewApiMember(endpoints)
	if err != nil {
		env.logger.Fatalln("Etcd.NewApiMember() found error:", err)
	}

	return memApi
}

// Watch restartMemberChannel
func clusterMemberRestartRoutine() {
	//flush last log info
	defer env.logger.Sync()

	for {
		select {
		case trigger := <-restartMemberChannel:
			if trigger == cluster.MEMBER_RESTART_CMDWAIT {
				env.logger.Println("exec restartMemberChannel MEMBER_RESTART_CMDWAIT...")
				// need to reset execCheckFailedTimes
				execCheckFailedTimes = 0

			} else if trigger == cluster.MEMBER_RESTART_HEALTHCHECK {
				env.logger.Println("exec restartMemberChannel MEMBER_RESTART_HEALTHCHECK...")
				if clusterCmd.Process != nil {
					env.logger.Println("Kill old process runtime, pid:", clusterCmd.Process.Pid)
					clusterCmd.Process.Kill()
				}

			} else {
				env.logger.Println("Fetch error restartMemberChannel value:", trigger)
			}

			//ProcessState stores information about a process, as reported by Wait.
			if clusterCmd.ProcessState != nil {
				env.logger.Println("Exception: clusterCmd.ProcessState is:", clusterCmd.ProcessState.String())
				bootstrapLocalClusterMember()

			} else {
				env.logger.Println("Exception: clusterCmd.ProcessState is nil.")
			}
		}
	}
}

func UpdateLatestEndpoints() {
	//flush last log info
	defer env.logger.Sync()

	memApi := getClientMembersApi(strings.Split(env.GetClientUrl(), ","))

	tmpEndpoints, err := memApi.GetInitialClusterEndpoints()
	if err != nil {
		env.logger.Fatalln("Etcd.GetInitialClusterEndpoints() found error:", err)
	}
	env.logger.Println("Fetch discovery service latest endpoints:", tmpEndpoints)

	// lock for update
	endpointsSyncLock.Lock()
	defer endpointsSyncLock.Unlock()
	lastestEndpoints = tmpEndpoints
}
