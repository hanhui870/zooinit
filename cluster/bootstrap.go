package cluster

import (
	"fmt"
	syslog "log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"

	"zooinit/cluster/etcd"
	"zooinit/config"
	"zooinit/log"
	"zooinit/utility"
)

const (
	CONFIG_SECTION            = "system.cluster"
	CLUSTER_BOOTSTRAP_TIMEOUT = 5 * time.Minute
	//change to 2, zookeeper default is 2 sec.
	CLUSTER_HEALTH_CHECK_INTERVAL = 2 * time.Second

	// 1. consul/config/size qurorum大小
	// 2. consul/election/ 候选人选举目录，CreateInOrder
	// 3. consul/members/ 正式集群中的候选人,需要health check 更新.
	CLUSTER_CONFIG_DIR = "/config"
	// 4. consul/config/size cluster qurorum size
	CLUSTER_CONFIG_DIR_SIZE = CLUSTER_CONFIG_DIR + "/size"
	// 5. consul/config/booted check whether the cluster is booted.
	CLUSTER_CONFIG_DIR_BOOTED = CLUSTER_CONFIG_DIR + "/booted"
	CLUSTER_SELECTION_DIR     = "/election"
	// 6. check health update this
	CLUSTER_MEMBER_DIR = "/members"
	// member node ttl
	CLUSTER_MEMBER_NODE_TTL = 1 * time.Minute

	//Cluster member restart channel value type
	MEMBER_RESTART_CMDWAIT     = 1
	MEMBER_RESTART_HEALTHCHECK = 2
	//Max failed will triger a restart cmd
	MEMBER_MAX_FAILED_TIMES = 10
)

var (
	env *envInfo

	// Discovery service latest endpoints
	lastestEndpoints  []string
	endpointsSyncLock sync.Mutex

	// Discovery service latest result members of election
	membersElected  []string
	membersSyncLock sync.Mutex

	// Election qurorum size
	qurorumSize       int
	qurorumWatchIndex uint64
	qurorumSyncLock   sync.Mutex

	// WaitGroup for goroutine to complete
	cmdWaitGroup sync.WaitGroup

	// Whether cluster is booted and healthy
	clusterUpdated bool
	// Where the cluster is booted before, for container rejoin
	clusterIsBootedBefore bool
	// Is terminate the app
	exitApp atomic.Value
	// Latest modify index for etcd qurorum watch
	latestIndex uint64

	callCmdStartInstance *exec.Cmd
	// Restart cluster member channel
	restartMemberChannel chan int
	execCheckFailedTimes int
)

func init() {
	// init channel
	restartMemberChannel = make(chan int)
	// false
	clusterIsBootedBefore = false
}

// Zooinit app runtime main line
func Bootstrap(c *cli.Context) {
	fname := config.GetConfigFileName(c.String("config"))
	iniobj := config.GetConfigInstance(fname)

	if len(c.Args()) != 1 {
		fmt.Println(c.Command.Usage)
		os.Exit(1)
	}

	service := c.Args()[0]
	service = strings.Trim(service, " \t\n\r")
	if service == "" {
		syslog.Fatalln("Command args of service name is empty.")
	} else if service == "bootstrap" {
		syslog.Fatalln("Service name of \"bootstrap\" is reserved.")
	}

	// backend of servie
	backend := c.String("backend")
	if backend == "" {
		backend = service
	}
	env = NewEnvInfo(iniobj, backend, service)

	//flush last log info
	defer env.logger.Sync()

	env.logger.Println("Logger path:", env.logPath)
	env.logger.Println("Timeout:", env.timeout.String())
	env.logger.Println("Qurorum:", env.qurorum)
	env.logger.Println("Discover method:", env.discoveryMethod)
	env.logger.Println("Discover path:", env.discoveryPath)
	env.logger.Println("env.discoveryTarget for fetch members:", env.discoveryTarget)

	// update endpoints
	UpdateLatestEndpoints()

	// register node
	initializeClusterDiscoveryInfo()

	// loop wait qurorum size of nodes is registed
	loopUntilQurorumIsReached()

	// start up local node
	bootstrapLocalClusterMember()

	// Will block
	loopUntilClusterIsUp(env.timeout)

	//must before watchDogRunning
	go clusterMemberRestartRoutine()

	// watch and check cluster health [watchdog], block until server receive term signal
	watchDogRunning()

	// final wait.
	cmdWaitGroup.Wait()
	env.logger.Println("App runtime reaches end.")
}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}

// init cluster bootstrap info
func initializeClusterDiscoveryInfo() {
	//flush last log info
	defer env.logger.Sync()

	kvApi := getClientKeysApi()

	// Set qurorum size
	resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath, "", &client.SetOptions{Dir: true, TTL: env.timeout, PrevExist: client.PrevNoExist})
	if err != nil {
		//ignore exist error
		if !etcd.EqualEtcdError(err, client.ErrorCodeNodeExist) {
			// check if exist need to add qurorum
			env.logger.Fatalln("Etcd.Api() set "+env.discoveryPath+" error:", err)
		} else {
			env.logger.Println("Etcd.Api() set " + env.discoveryPath + " notice: node exist, will add qurorum directly.")

			resp, err = kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED, &client.GetOptions{})
			if err == nil {
				env.logger.Println("Etcd.Api() cluster already booted at :", CLUSTER_CONFIG_DIR_BOOTED, ", will check whether in the nodelist...")

				// fetch latest node list
				nodeList, err := getLastestNodeList()
				if err != nil {
					env.logger.Fatalln("Etcd.Api() get "+env.discoveryPath+CLUSTER_SELECTION_DIR+" lastest election nodes error:", err)
				} else {
					if utility.InSlice(nodeList, env.localIP.String()) {
						env.logger.Println("Etcd.Api() local machine is in the nodelist:", env.localIP.String(), nodeList, ", will continue to restart...")

						clusterIsBootedBefore = true
					} else {
						env.logger.Fatalln("Etcd.Api() local machine is NOT in the nodelist:", env.localIP.String(), nodeList)
					}
				}

			} else if !etcd.EqualEtcdError(err, client.ErrorCodeKeyNotFound) {
				env.logger.Fatalln("Etcd.Api() found error while fetch:", CLUSTER_CONFIG_DIR_BOOTED, " error:", err)
			}
		}

	} else {
		// Create success, set config
		env.logger.Println("Etcd.Api() set "+env.discoveryPath+" ok, TTL:", env.timeout.String(), resp)

		// Set config size
		resp, err = kvApi.Conn().Set(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_SIZE, strconv.Itoa(env.qurorum), &client.SetOptions{PrevExist: client.PrevNoExist})
		if err != nil {
			env.logger.Fatalln("Etcd.Api() set "+CLUSTER_CONFIG_DIR_SIZE+" error:", err)
		} else {
			env.logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_SIZE+" ok, Qurorum size:", env.qurorum, resp)
		}
	}

	// Call script
	if env.eventOnPreRegist != "" {
		callCmd := getCallCmdInstance("OnPreRegist: ", env.eventOnPreRegist)
		cmdCallWaitProcessSync(callCmd)
	}

	// Create qurorum in order node
	resp, err = kvApi.Conn().CreateInOrder(etcd.Context(), env.discoveryPath+CLUSTER_SELECTION_DIR, env.localIP.String(), nil)
	if err != nil {
		env.logger.Fatalln("Etcd.Api() CreateInOrder error:", err)
	} else {
		env.logger.Println("Etcd.Api() CreateInOrder ok:", resp)

		// init create in order latestIndex
		latestIndex = resp.Node.CreatedIndex - 1

		// Call script
		if env.eventOnPostRegist != "" {
			callCmd := getCallCmdInstance("OnPostRegist: ", env.eventOnPostRegist)
			cmdCallWaitProcessSync(callCmd)
		}
	}

	// Finish
}

func UpdateLatestEndpoints() {
	//flush last log info
	defer env.logger.Sync()

	memApi := getClientMembersApi(strings.Split(env.discoveryTarget, ","))

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

func loopUntilQurorumIsReached() {
	//flush last log info
	defer env.logger.Sync()

	kvApi := getClientKeysApi()

	// GetLatestElectionMember index incrs from this one
	if !clusterIsBootedBefore {
		resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_SELECTION_DIR, &client.GetOptions{Recursive: true, Sort: true})
		if err != nil {
			env.logger.Fatalln("Etcd.Api() get "+env.discoveryPath+CLUSTER_SELECTION_DIR+" lastest ModifiedIndex error:", err)

		} else {
			latestIndex = resp.Node.ModifiedIndex - 1 // latestIndex-1 for check unitial change
		}
	}

	// GetConfigSize
	getQurorumSize()
	// Change concurrently
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		watchQurorumSize()
	}()

	// loop until qurorum size is reached
	for {
		wather := kvApi.Conn().Watcher(env.discoveryPath+CLUSTER_SELECTION_DIR, &client.WatcherOptions{Recursive: true, AfterIndex: latestIndex})
		resp, err := wather.Next(etcd.Context())
		if err != nil {
			env.logger.Fatalln("Etcd.Watcher() watch "+env.discoveryPath+CLUSTER_SELECTION_DIR+" error:", err)

		} else {
			latestIndex = resp.Node.ModifiedIndex
			env.logger.Println("Get last ModifiedIndex of watch:", latestIndex)
		}

		// fetch latest node list
		nodeList, err := getLastestNodeList()
		if err != nil {
			env.logger.Println("Etcd.Api() get "+env.discoveryPath+CLUSTER_SELECTION_DIR+" lastest election nodes error:", err)
			env.logger.Println("Will exit now...")

			//Need to expecility exit
			//Sleep for finish action at watchQurorumSize()
			for {
				if isExit, ok := exitApp.Load().(bool); ok && isExit {
					env.logger.Println("Wait completed, bye...")
					os.Exit(0)
				}

				time.Sleep(100 * time.Millisecond)
			}

		} else {
			env.logger.Println("Get election qurorum size, after remove duplicates:", len(nodeList), "members:", nodeList)

			if len(nodeList) >= qurorumSize {
				membersSyncLock.Lock()
				membersElected = nodeList[:qurorumSize]
				membersSyncLock.Unlock()

				env.logger.Println("Get election qurorum finished:", membersElected)

				// Call script
				if env.eventOnReachQurorumNum != "" {
					callCmd := getCallCmdInstance("OnReachQurorumNum: ", env.eventOnReachQurorumNum)
					cmdCallWaitProcessSync(callCmd)
				}
				break
			}
		}
	}
}

func getLastestNodeList() ([]string, error) {
	//flush last log info
	defer env.logger.Sync()

	kvApi := getClientKeysApi()

	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_SELECTION_DIR, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, err
	} else {
		var nodeList []string
		for _, node := range resp.Node.Nodes {
			if node.Dir || !etcd.CheckInOrderKeyFormat(node.Key) {
				continue
			}
			nodeList = append(nodeList, node.Value)
		}

		nodeList = utility.RemoveDuplicateInOrder(nodeList)
		return nodeList, nil
	}
}

func bootstrapLocalClusterMember() {
	//flush last log info
	defer env.logger.Sync()

	env.logger.Println("Started to boot Local cluster member, local ip:", env.localIP.String())

	if !utility.InSlice(membersElected, env.localIP.String()) {
		env.logger.Fatalln("BootstrapLocalClusterMember error, localip is not in the elected list.")
	}

	// Call script block
	if env.eventOnPreStart != "" {
		callCmd := getCallCmdInstance("OnPreStart: ", env.eventOnPreStart)
		cmdCallWaitProcessSync(callCmd)
	}

	// Call script, non block
	callCmdStartInstance = getCallCmdInstance("OnStart: ", env.eventOnStart)
	err := callCmdStartInstance.Start()
	if err != nil {
		env.logger.Println("callCmd.Start() error found:", err)
	}

	// block until cluster is up
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		err = callCmdStartInstance.Wait()
		if err != nil {
			env.logger.Println("callCmd.Wait() finished with error found:", err)
		} else {
			env.logger.Println("callCmd.Wait() finished without error.")
		}

		if isExit, ok := exitApp.Load().(bool); !ok || !isExit {
			env.logger.Println("BootstrapLocalClusterMember do not detect exitApp cmd, will restart...")
			restartMemberChannel <- MEMBER_RESTART_CMDWAIT
		}
	}()
}

func getCallCmdInstance(logPrefix, event string) *exec.Cmd {
	callCmd := exec.Command("bash", "-c", "script/main.py")

	loggerIOAdapter := log.NewLoggerIOAdapter(env.logger)
	loggerIOAdapter.SetPrefix(logPrefix)
	callCmd.Stdout = loggerIOAdapter
	callCmd.Stderr = loggerIOAdapter

	// Transfer variable to python
	callCmd.Env = getCallCmdENVSet(event)

	return callCmd
}

// Async wait by WaitGroup
func cmdCallWaitProcessAsync(callCmd *exec.Cmd) {
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		err := callCmd.Start()
		if err != nil {
			env.logger.Println("callCmd.Start() error found:", err)
		}
		callCmd.Wait()
	}()
}

// Sync block wait
func cmdCallWaitProcessSync(callCmd *exec.Cmd) {
	defer func() {
		if callCmd.Process != nil {
			callCmd.Process.Kill()
		}
	}()

	err := callCmd.Start()
	if err != nil {
		env.logger.Println("callCmd.Start() error found:", err)
	}
	callCmd.Wait()
}

func getCallCmdENVSet(event string) []string {

	envs := []string{"ZOOINIT_CLUSTER_BACKEND=" + env.clusterBackend}
	envs = append(envs, "ZOOINIT_CLUSTER_SERVICE="+env.service)
	envs = append(envs, "ZOOINIT_CLUSTER_EVENT="+event)
	envs = append(envs, "ZOOINIT_SERVER_IP_LIST="+strings.Join(membersElected, ","))
	envs = append(envs, "ZOOINIT_LOCAL_IP="+env.localIP.String())

	// Master first one
	if len(membersElected) > 0 {
		envs = append(envs, "ZOOINIT_MASTER_IP="+membersElected[0])
	} else {
		envs = append(envs, "ZOOINIT_MASTER_IP=")
	}

	//defalut 0
	envs = append(envs, "ZOOINIT_QURORUM="+strconv.Itoa(qurorumSize))

	//need to sync PATH ENV
	envs = append(envs, "PATH="+os.Getenv("PATH"))

	return envs
}

// Cluster stated check may also need to use scripts hook
func loopUntilClusterIsUp(timeout time.Duration) (result bool, err error) {
	//flush last log info
	defer env.logger.Sync()

	kvApi := getClientKeysApi()

	result = false

	// Important!!! check upstarted
	env.logger.Println("Call hook script for check discovery cluster's startup...")
	// Call script
	cmdWaitGroup.Add(1)
	sucCh := make(chan bool)
	go func() {
		defer cmdWaitGroup.Done()

		for {
			//Need init every call: callCmd.Start() error found: exec: already started
			callCmd := getCallCmdInstance("OnPostStart: ", env.eventOnPostStart)

			err := callCmd.Start()
			if err != nil {
				env.logger.Println("callCmd.Start() error found:", err)
			}
			callCmd.Wait()

			// Check cluster if up if return exit code is normal
			if callCmd.ProcessState.Success() {
				sucCh <- true

				env.logger.Println("Cluster is checked up now, The status is normal.")
				clusterUpdated = true

				// schedule to update discovery path ttl
				go updateDiscoveryTTL()

				// Set config/booted to true
				booted := "true," + env.localIP.String() + "," + time.Now().String()
				resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED, booted, &client.SetOptions{PrevExist: client.PrevNoExist})
				if err != nil {
					//ignore exist error
					if etcd.EqualEtcdError(err, client.ErrorCodeNodeExist) {
						// check if exist need to add qurorum
						env.logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED+" set by another node, error:", err)

						resp, err = kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED, &client.GetOptions{})
						if err == nil {
							env.logger.Println("Etcd.Api() cluster already booted at :", CLUSTER_CONFIG_DIR_BOOTED, "Resp:", resp)
						} else {
							env.logger.Fatalln("Etcd.Api() found error while fetch:", CLUSTER_CONFIG_DIR_BOOTED, " error:", err)
						}
					} else {
						env.logger.Fatalln("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED+" error:", err)
					}
				} else {
					env.logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED+" ok:", booted, "Resp:", resp)
				}
				break
			}
		}
	}()

	select {
	case <-time.After(env.timeout):
		env.logger.Println("Cluster booting is timeout, will give up and terminate.")
		exitApp.Store(true)
	case <-sucCh:
		break
	}

	return result, err
}

// usage: go updateDiscoveryTTL()
// update need to call after
func updateDiscoveryTTL() {
	//flush last log info
	defer env.logger.Sync()
	rand.Seed(time.Now().UnixNano())

	//30-59s 随机更新一次,所有节点都会更新
	for {
		kvApi := getClientKeysApi()

		// Update TTL
		resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath, "", &client.SetOptions{Dir: true, TTL: env.timeout, PrevExist: client.PrevExist})
		if err != nil {
			env.logger.Fatalln("Etcd.Api() update "+env.discoveryPath+" TTL error:", err)
		} else {
			env.logger.Println("Etcd.Api() update "+env.discoveryPath+" ok", "Resp:", resp)
		}

		next := time.Duration(30 + rand.Intn(30))
		time.Sleep(next * time.Second)
	}
}

func watchDogRunning() {
	//flush last log info
	defer env.logger.Sync()

	// Call OnClusterBooted script, no need goroutine
	if env.eventOnClusterBooted != "" {
		callCmd := getCallCmdInstance("OnClusterBooted: ", env.eventOnClusterBooted)
		err := callCmd.Start()
		if err != nil {
			env.logger.Println("callCmd.Start() error found:", err)
		}
		callCmd.Wait()
	}

	//this can not use goroutine, this is a loop
	firstRun := true
	for {
		if isExit, ok := exitApp.Load().(bool); ok && isExit {
			env.logger.Println("Receive exitApp signal, break watchDogRunning loop.")
			break
		}

		execHealthChechRunning(firstRun)
		//do not need break, because loop is maitained by zooinit

		if firstRun {
			firstRun = false
		}

		// sleep interval time
		time.Sleep(env.healthCheckInterval)
	}
}

// Check cluster health after cluster is up.
func execHealthChechRunning(firstRun bool) (result bool) {
	//flush last log info
	defer env.logger.Sync()

	callCmd := getCallCmdInstance("OnHealthCheck: ", env.eventOnHealthCheck)
	if !firstRun {
		// like exec health check, no need to debug env info every time.
		callCmd.Env = append(callCmd.Env, "ZOOINIT_SILENT_ENV_INFO=true")
	}

	// may runtime error: invalid memory address or nil pointer dereference
	cmdCallWaitProcessSync(callCmd)

	var cm *ClusterMember
	// ttl 1min, update 1/s
	if callCmd.ProcessState != nil && callCmd.ProcessState.Success() {
		// when the health check call normal return, break the infinite loop
		result = true
		// reset to 0
		execCheckFailedTimes = 0
	} else {
		result = false

		// trigger restart related
		execCheckFailedTimes++
		if execCheckFailedTimes >= MEMBER_MAX_FAILED_TIMES {
			env.logger.Println("Cluster member is NOT healthy, will trigger Restart. Failed times:", execCheckFailedTimes, ", MEMBER_MAX_FAILED_TIMES:", MEMBER_MAX_FAILED_TIMES)
			execCheckFailedTimes = 0
			restartMemberChannel <- MEMBER_RESTART_HEALTHCHECK
		} else {
			env.logger.Println("Cluster member is NOT healthy, Failed times:", execCheckFailedTimes)
		}
	}

	cm = NewClusterMember(env.GetNodename(), env.localIP.String(), result, execCheckFailedTimes)
	kvApi := getClientKeysApi()
	pathNode := env.discoveryPath + CLUSTER_MEMBER_DIR + "/" + env.GetNodename()
	resp, err := kvApi.Conn().Set(etcd.Context(), pathNode, cm.ToJson(), &client.SetOptions{Dir: false, TTL: CLUSTER_MEMBER_NODE_TTL})
	if err != nil {
		env.logger.Fatalln("Etcd.Api() update "+pathNode+" State error:", err)
	} else {
		env.logger.Println("Etcd.Api() update "+pathNode+" ok", "Resp:", resp)
	}

	return result
}

// Need to watch config size
func watchQurorumSize() {
	//flush last log info
	defer env.logger.Sync()

	for {
		kvApi := getClientKeysApi()
		watch := kvApi.Conn().Watcher(env.discoveryPath+CLUSTER_CONFIG_DIR_SIZE, &client.WatcherOptions{AfterIndex: qurorumWatchIndex})
		resp, err := watch.Next(etcd.Context())

		// if Cluster is booted, quit.
		if clusterUpdated {
			break
		}

		if err != nil {
			env.logger.Fatalln("Etcd.Api() watch "+CLUSTER_CONFIG_DIR_SIZE+" error:", err)
		} else {
			env.logger.Println("Etcd.Api() watch "+CLUSTER_CONFIG_DIR_SIZE+" found change:", resp)

			if (resp.Action == "expire" || resp.Action == "delete") && resp.Node.Key == env.discoveryPath {
				env.logger.Println("Etcd.Api() service boot timeout reach, will delete " + env.discoveryPath + " and terminate app.")

				resp, err := kvApi.Conn().Delete(etcd.Context(), env.discoveryPath, &client.DeleteOptions{Recursive: true, Dir: true})

				if err != nil {
					env.logger.Println("Etcd.Api() error while delete "+env.discoveryPath+": ", err)
				} else {
					env.logger.Println("Etcd.Api() deleted "+env.discoveryPath+" and bye: ", resp)
				}

				// Need exit
				exitApp.Store(true)
			} else {
				getQurorumSize()
			}
		}
	}
}

func getQurorumSize() {
	//flush last log info
	defer env.logger.Sync()

	kvApi := getClientKeysApi()
	// get config size
	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_SIZE, &client.GetOptions{})
	if err != nil {
		env.logger.Fatalln("Etcd.Api() get "+CLUSTER_CONFIG_DIR_SIZE+" error:", err)
	} else {
		env.logger.Println("Etcd.Api() get "+CLUSTER_CONFIG_DIR_SIZE+" ok, Qurorum size:", resp.Node.Value)
		tmp, err := strconv.ParseInt(resp.Node.Value, 10, 64)
		if err != nil {
			env.logger.Fatalln("Error: strconv.ParseInt("+CLUSTER_CONFIG_DIR_SIZE+") error:", err)
		}

		qurorumSyncLock.Lock()
		qurorumSize = int(tmp)
		if latestIndex > resp.Node.ModifiedIndex {
			qurorumWatchIndex = latestIndex
		} else {
			qurorumWatchIndex = resp.Node.ModifiedIndex
		}

		qurorumSyncLock.Unlock()
	}
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
			if trigger == MEMBER_RESTART_CMDWAIT {
				env.logger.Println("exec restartMemberChannel MEMBER_RESTART_CMDWAIT...")
				// need to reset execCheckFailedTimes
				execCheckFailedTimes = 0

			} else if trigger == MEMBER_RESTART_HEALTHCHECK {
				env.logger.Println("exec restartMemberChannel MEMBER_RESTART_HEALTHCHECK...")
				if callCmdStartInstance.Process != nil {
					env.logger.Println("Kill old process runtime, pid:", callCmdStartInstance.Process.Pid)
					callCmdStartInstance.Process.Kill()
				}

			} else {
				env.logger.Println("Fetch error restartMemberChannel value:", trigger)
			}

			//ProcessState stores information about a process, as reported by Wait.
			if callCmdStartInstance.ProcessState != nil {
				env.logger.Println("Exception: callCmdStartInstance.ProcessState is:", callCmdStartInstance.ProcessState.String())
				bootstrapLocalClusterMember()

			} else {
				env.logger.Println("Exception: callCmdStartInstance.ProcessState is nil.")
			}
		}
	}
}
