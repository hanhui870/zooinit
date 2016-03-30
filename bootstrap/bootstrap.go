package bootstrap

import (
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"

	"zooinit/cluster"
	"zooinit/cluster/etcd"
	"zooinit/config"
	"zooinit/log"
	"zooinit/utility"
)

const (
	CONFIG_SECTION = "system.boostrap"

	// INTERNAL discovery findpath
	INTERNAL_FINDPATH         = "/zooinit/boot"
	CLUSTER_BOOTSTRAP_TIMEOUT = 5 * time.Minute

	DEFAULT_BOOTSTRAP_DISCOVERY_PATH = "/zooinit/discovery/cluster/" + cluster.BOOTSTRAP_SERVICE_NAME
)

var (
	env *envInfo

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

	// Whether cluster is booted and healthy
	clusterUpdated bool
)

func init() {
	// init channel
	restartMemberChannel = make(chan int)
	clusterUpdated = false
}

func Bootstrap(c *cli.Context) {
	fname := config.GetConfigFileName(c.String("config"))
	iniobj := config.GetConfigInstance(fname)

	env = NewEnvInfo(iniobj, c)

	cluster.GuaranteeSingleRun(env)

	//register signal watcher
	env.RegisterSignalWatch()

	BootstrapEtcd(env)
}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}

func BootstrapEtcd(env *envInfo) {
	// flush last log info
	defer env.Logger.Sync()
	env.Logger.Println("Starting to boot Etcd...")

	env.Logger.Println("Logger channel:", env.LogChannel)
	if env.LogChannel != log.LOG_STDOUT {
		env.Logger.Println("Logger path:", env.LogPath)
	}
	env.Logger.Println("Timeout:", env.Timeout.String())
	env.Logger.Println("Qurorum:", env.Qurorum)
	env.Logger.Println("Discover internal:", env.GetInternalClientUrl())
	env.Logger.Println("Health check interval:", env.HealthCheckInterval)

	// Boot Internal etcd
	if env.isSelfIp {
		bootUpInternalEtcd()
	}

	// Boot local cluster member
	bootstrapLocalClusterMember()

	//must before watchDogRunning, can before cluster is up.
	go clusterMemberRestartRoutine()

	checkDiscoveryClusterIsUp()

	// Up and fetch latest endpoints.
	UpdateLatestEndpoints()

	// watch and check cluster health [watchdog], block until server receive term signal
	// check cluster bootstraped and register memberself
	// If stoped, process's output can't trace no longer
	watchDogRunning()

	// final wait.
	cmdWaitGroup.Wait()

	env.Logger.Println("App runtime reaches end.")
}

func bootUpInternalEtcd() {
	// flush last log info
	defer env.Logger.Sync()

	// Api to internal service
	api, err := etcd.NewApiKeys([]string{env.GetInternalClientUrl()})
	if err != nil {
		env.Logger.Fatal("Etcd NewApi error:", err)
	}

	env.Logger.Println("Etcd Internal PeerUrl:", env.GetInternalPeerUrl())
	env.Logger.Println("Etcd Internal ClientUrl:", env.GetInternalClientUrl())

	// Add & can't fast wait
	// data-dir can't be same with discovery service.
	intName := "etcd.initial"
	intExecCmd := env.cmd + " --data-dir " + env.internalDataDir + " -wal-dir " + env.internalWalDir + " -name " + intName +
		" -initial-advertise-peer-urls " + env.GetInternalPeerUrl() +
		" -listen-peer-urls " + env.GetInternalPeerUrl() +
		" -listen-client-urls " + env.GetInternalClientUrl() +
		" -advertise-client-urls " + env.GetInternalClientUrl() +
		" -initial-cluster " + intName + "=" + env.GetInternalPeerUrl()

	env.Logger.Println("Etcd Internal ExecCmd:", intExecCmd)

	// Boot internal discovery service
	path, args, err := utility.ParseCmdStringWithParams(intExecCmd)
	if err != nil {
		env.Logger.Fatalln("Error ParseCmdStringWithParams internal service:", err)
	}

	internalCmd = exec.Command(path, args...)
	loggerIOAdapter := log.NewLoggerIOAdapter(env.Logger)
	loggerIOAdapter.SetPrefix("Etcd internal server: ")
	internalCmd.Stdout = loggerIOAdapter
	internalCmd.Stderr = loggerIOAdapter
	err = internalCmd.Start()

	// internal cmd wait
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		//promise kill sub process
		defer func() {
			if internalCmd.Process != nil {
				internalCmd.Process.Kill()
			}
		}()

		err = internalCmd.Wait()
		if err != nil {
			env.Logger.Println("internalCmd.Wait() finished with error found:", err)
		} else {
			env.Logger.Println("internalCmd.Wait() finished without error.")
		}
	}()

	if err != nil {
		env.Logger.Fatalln("Exec Internal ExecCmd Error:", err)
	} else {
		env.Logger.Println("Exec Internal OK, PID:", internalCmd.Process.Pid)

		// Set PID
		env.internalCmdInstance = internalCmd
		env.Logger.Println("Internal service started.")

		// Important!!! check upstarted
		env.Logger.Println("Etcd LoopTimeoutRequest for check internal's startup...")

		internalCheckout := 3 * time.Second
		isHealth, err := LoopTimeoutRequest(internalCheckout, env, func() (bool, error) {
			return etcd.CheckHealth(env.GetInternalClientUrl())
		})
		if err != nil {
			env.Logger.Fatal("Error check internal error: ", err)
		} else if isHealth != true {
			env.Logger.Fatal("Error check internal server health: ", isHealth)
			env.Logger.Fatal("Cluster bootstrap faild: failed to bootstrap in ", internalCheckout.String())
		}

		resp, err := http.Get(env.GetInternalClientUrl() + "/v2/stats/self")
		if err != nil {
			env.Logger.Fatal("Error fetch stats self: ", err)
		}
		env.Logger.Println("Etcd internal Stat self: ", resp)

		_, err = api.Conn().Delete(etcd.Context(), INTERNAL_FINDPATH, &client.DeleteOptions{Dir: true, Recursive: true})
		if err != nil {
			//type safe cast
			err, ok := err.(client.Error)
			if ok && err.Code != client.ErrorCodeKeyNotFound {
				env.Logger.Fatal("Delete ", INTERNAL_FINDPATH, " error:", err)
			}
		}

		env.Logger.Println("Set Cluster bootstrap timeout:", env.Timeout.String())
		_, err = api.Conn().Set(etcd.Context(), INTERNAL_FINDPATH, "", &client.SetOptions{TTL: env.Timeout, Dir: true})
		if err != nil {
			env.Logger.Fatal("Set TTL for ", INTERNAL_FINDPATH, " error:", err)
		}

		env.Logger.Println("Set Qurorum ", INTERNAL_FINDPATH+"/_config/size to ", env.Qurorum)
		_, err = api.Conn().Set(etcd.Context(), INTERNAL_FINDPATH+"/_config/size", strconv.Itoa(env.Qurorum), nil)
		if err != nil {
			env.Logger.Fatal("Set Qurorum ", INTERNAL_FINDPATH+"/_config/size error: ", err)
		}
	}
}

func bootstrapLocalClusterMember() {
	//flush last log info
	defer env.Logger.Sync()

	// Cluster member startup info
	env.Logger.Println("Etcd Discovery PeerUrl:", env.GetPeerUrl())
	env.Logger.Println("Etcd Discovery ClientUrl:", env.GetClientUrl())

	// Etcd cluster can restart, because etcd restart don't need discovery service
	disExecCmd := env.cmd + " --data-dir " + env.cmdDataDir + " -wal-dir " + env.cmdWalDir +
		" -snapshot-count " + strconv.Itoa(env.cmdSnapCount) +
		" -name " + "etcd.bootstrap." + env.LocalIP.String() +
		" -initial-advertise-peer-urls " + env.GetPeerUrl() +
		" -listen-peer-urls " + env.GetPeerUrl() +
		" -listen-client-urls http://127.0.0.1:2379," + env.GetClientUrl() +
		" -advertise-client-urls " + env.GetClientUrl() +
		" -discovery " + env.GetInternalClientUrl() + "/v2/keys" + INTERNAL_FINDPATH

	env.Logger.Println("Etcd Discovery ExecCmd: ", disExecCmd)

	// Boot internal discovery service
	// Need to rm -rf /tmp/etcd/ because dir may be used before
	path, args, err := utility.ParseCmdStringWithParams(disExecCmd)
	if err != nil {
		env.Logger.Fatalln("Error ParseCmdStringWithParams cluster bootstrap: ", err)
	}

	clusterCmd = exec.Command(path, args...)
	loggerIOAdapter := log.NewLoggerIOAdapter(env.Logger)
	loggerIOAdapter.SetPrefix("Etcd discovery member: ")
	clusterCmd.Stdout = loggerIOAdapter
	clusterCmd.Stderr = loggerIOAdapter

	err = clusterCmd.Start()
	if err != nil {
		env.Logger.Fatalln("Exec Discovery ExecCmd Error: ", err)
	} else {
		env.Logger.Println("Exec Discovery Etcd member OK, PID: ", clusterCmd.Process.Pid)
		env.Logger.Println("Etcd member service ", env.GetClientUrl(), " started,  waiting to be bootrapped.")
	}

	// here no block now, block in health check
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		//promise kill sub process
		defer func() {
			if clusterCmd.Process != nil {
				clusterCmd.Process.Kill()
			}
		}()

		err = clusterCmd.Wait()
		if err != nil {
			env.Logger.Println("callCmd.Wait() finished with error found:", err)
		} else {
			env.Logger.Println("callCmd.Wait() finished without error.")
		}

		if isExit, ok := exitApp.Load().(bool); !ok || !isExit {

			env.Logger.Println("BootstrapLocalClusterMember do not detect exitApp cmd, will restart...")
			restartMemberChannel <- cluster.MEMBER_RESTART_CMDWAIT
		}
	}()
}

func checkDiscoveryClusterIsUp() {
	//flush last log info
	defer env.Logger.Sync()

	// Important!!! check upstarted
	env.Logger.Println("Etcd LoopTimeoutRequest for check discovery cluster's startup...")
	isHealth, err := LoopTimeoutRequest(env.Timeout, env, func() (bool, error) {
		return etcd.CheckHealth(env.GetClientUrl())
	})
	if err != nil {
		env.Logger.Fatal("Error check discovery error: ", err)
	} else if isHealth != true {
		env.Logger.Fatal("Error check discovery server health: ", isHealth)
		env.Logger.Fatal("Cluster bootstrap faild: failed to bootstrap in ", env.Timeout.String())
	}

	clusterUpdated = true

	// Close internal service
	env.Logger.Println("Cluster etcd service is booted. Internal service is going to be terminated.")
	if internalCmd != nil && internalCmd.Process != nil {
		internalCmd.Process.Kill()
	}
}

//request until sucess
func LoopTimeoutRequest(timeout time.Duration, env *envInfo, routine func() (result bool, err error)) (result bool, err error) {
	var charlist []byte

	//flush last log info
	defer env.Logger.Sync()

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

	env.Logger.Println("Fetched data LoopTimeoutRequest for loop:", string(charlist))

	return result, err
}

func watchDogRunning() {
	//flush last log info
	defer env.Logger.Sync()

	//failedTimes update to etcd
	failedTimes := 0
	for {
		if isExit, ok := exitApp.Load().(bool); ok && isExit {
			env.Logger.Println("Receive exitApp signal, break watchDogRunning loop.")
			break
		}

		var cm *cluster.ClusterMember
		var result bool
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
				env.Logger.Println("Found error while etcd.CheckHealth("+env.GetClientUrl()+"):", err)
			}

			if execCheckFailedTimes >= cluster.MEMBER_MAX_FAILED_TIMES {
				env.Logger.Println("Cluster member is NOT healthy, will trigger Restart. Failed times:", execCheckFailedTimes, ", MEMBER_MAX_FAILED_TIMES:", cluster.MEMBER_MAX_FAILED_TIMES)
				execCheckFailedTimes = 0
				restartMemberChannel <- cluster.MEMBER_RESTART_HEALTHCHECK
			} else {
				env.Logger.Println("Cluster member is NOT healthy, Failed times:", execCheckFailedTimes)
			}
		}

		kvApi := getClientKeysApi()
		cm = cluster.NewClusterMember(env.GetNodename(), env.LocalIP.String(), result, execCheckFailedTimes)
		pathNode := DEFAULT_BOOTSTRAP_DISCOVERY_PATH + cluster.CLUSTER_MEMBER_DIR + "/" + env.GetNodename()
		resp, err := kvApi.Conn().Set(etcd.Context(), pathNode, cm.ToJson(), &client.SetOptions{Dir: false, TTL: cluster.CLUSTER_MEMBER_NODE_TTL})
		if err != nil {
			env.Logger.Println("Etcd.Api() update "+pathNode+" State error:", err, " faildTimes:", failedTimes)
		} else {
			env.Logger.Println("Etcd.Api() update "+pathNode+" ok", "Resp:", resp)
			failedTimes = 0
		}

		// sleep interval time
		time.Sleep(env.HealthCheckInterval)
	}
}

func getClientKeysApi() *etcd.ApiKeys {
	//flush last log info
	defer env.Logger.Sync()

	kvApi, err := etcd.NewApiKeys(lastestEndpoints)
	if err != nil {
		env.Logger.Fatalln("Etcd.NewApiKeys() found error:", err)
	}

	return kvApi
}

func getClientMembersApi(endpoints []string) *etcd.ApiMembers {
	//flush last log info
	defer env.Logger.Sync()

	memApi, err := etcd.NewApiMember(endpoints)
	if err != nil {
		env.Logger.Fatalln("Etcd.NewApiMember() found error:", err)
	}

	return memApi
}

// Watch restartMemberChannel
func clusterMemberRestartRoutine() {
	//flush last log info
	defer env.Logger.Sync()

	restartTimes := 0
	for {
		select {
		case trigger := <-restartMemberChannel:
			if trigger == cluster.MEMBER_RESTART_CMDWAIT {
				env.Logger.Println("exec restartMemberChannel MEMBER_RESTART_CMDWAIT...")
				// need to reset execCheckFailedTimes
				execCheckFailedTimes = 0

			} else if trigger == cluster.MEMBER_RESTART_HEALTHCHECK {
				env.Logger.Println("exec restartMemberChannel MEMBER_RESTART_HEALTHCHECK...")
				if clusterCmd.Process != nil {
					env.Logger.Println("Kill old process runtime, pid:", clusterCmd.Process.Pid)
					clusterCmd.Process.Kill()
				}

			} else {
				env.Logger.Println("Fetch error restartMemberChannel value:", trigger)
			}

			// restart frequency control
			if restartTimes > 60 {
				restartTimes = 1 // reset
			}
			timeToSleep := time.Duration(restartTimes) * time.Second
			env.Logger.Println("Will sleep " + timeToSleep.String() + " continue to restart member...")
			time.Sleep(timeToSleep)
			restartTimes++

			//ProcessState stores information about a process, as reported by Wait.
			if clusterCmd.ProcessState != nil {
				env.Logger.Println("Exception: clusterCmd.ProcessState is:", clusterCmd.ProcessState.String())
				bootstrapLocalClusterMember()

			} else {
				env.Logger.Println("Exception: clusterCmd.ProcessState is nil.")
			}
		}
	}
}

func UpdateLatestEndpoints() {
	//flush last log info
	defer env.Logger.Sync()

	memApi := getClientMembersApi(strings.Split(env.GetClientUrl(), ","))

	tmpEndpoints, err := memApi.GetInitialClusterEndpoints()
	if err != nil {
		env.Logger.Fatalln("Etcd.GetInitialClusterEndpoints() found error:", err)
	}
	env.Logger.Println("Fetch discovery service latest endpoints:", tmpEndpoints)

	// lock for update
	endpointsSyncLock.Lock()
	defer endpointsSyncLock.Unlock()
	lastestEndpoints = tmpEndpoints
}
