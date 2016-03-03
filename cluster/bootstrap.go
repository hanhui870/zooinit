package cluster

import (
	"fmt"
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

	// 1. consul/config/size qurorum大小
	// 2. consul/election/ 候选人选举目录，CreateInOrder
	// 3. consul/members/ 正式集群中的候选人
	CLUSTER_CONFIG_DIR    = "/config"
	CLUSTER_SELECTION_DIR = "/election"
	CLUSTER_MEMBER_DIR    = "/members"
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
	// Is terminate the app
	exitApp atomic.Value
)

// Zooinit app runtime main line
func Bootstrap(c *cli.Context) {
	fname := config.GetConfigFileName(c.String("config"))
	iniobj := config.GetConfigInstance(fname)

	if len(c.Args()) != 1 {
		fmt.Println(c.Command.Usage)
		os.Exit(1)
	}

	cluster := c.Args()[0]
	env = NewEnvInfo(iniobj, cluster)

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
		}

	} else {
		// Create success, set config
		env.logger.Println("Etcd.Api() set "+env.discoveryPath+" ok, TTL:", env.timeout.String(), resp)

		// Set config size
		resp, err = kvApi.Conn().Set(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR+"/size", strconv.Itoa(env.qurorum), &client.SetOptions{PrevExist: client.PrevNoExist})
		if err != nil {
			env.logger.Fatalln("Etcd.Api() set /config/size error:", err)
		} else {
			env.logger.Println("Etcd.Api() set /config/size ok, Qurorum size:", env.qurorum, resp)
		}
	}

	// Call script
	if env.eventOnPreRegist != "" {
		callCmd := getCallCmdInstance("OnPreRegist: ", env.eventOnPreRegist)
		cmdCallWaitProcess(callCmd)
	}

	// Create qurorum in order node
	resp, err = kvApi.Conn().CreateInOrder(etcd.Context(), env.discoveryPath+CLUSTER_SELECTION_DIR, env.localIP.String(), nil)
	if err != nil {
		env.logger.Fatalln("Etcd.Api() CreateInOrder error:", err)
	} else {
		env.logger.Println("Etcd.Api() CreateInOrder ok:", resp)

		// Call script
		if env.eventOnPostRegist != "" {
			callCmd := getCallCmdInstance("OnPostRegist: ", env.eventOnPostRegist)
			cmdCallWaitProcess(callCmd)
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

	var latestIndex uint64
	// GetLatestElectionMember index incrs from this one
	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_SELECTION_DIR, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		env.logger.Fatalln("Etcd.Api() get "+env.discoveryPath+CLUSTER_SELECTION_DIR+" lastest ModifiedIndex error:", err)

	} else {
		latestIndex = resp.Node.ModifiedIndex - 1 // latestIndex-1 for check unitial change
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
		resp, err = wather.Next(etcd.Context())
		if err != nil {
			env.logger.Fatalln("Etcd.Watcher() watch "+env.discoveryPath+CLUSTER_SELECTION_DIR+" error:", err)

		} else {
			latestIndex = resp.Node.ModifiedIndex
			env.logger.Println("Get last ModifiedIndex of watch:", latestIndex)
		}

		resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_SELECTION_DIR, &client.GetOptions{Recursive: true, Sort: true})
		if err != nil {
			env.logger.Println("Etcd.Api() get "+env.discoveryPath+CLUSTER_SELECTION_DIR+" lastest election nodes error:", err)
			env.logger.Println("Will exit now...")
			//Sleep for finish action at watchQurorumSize()
			time.Sleep(time.Second)

		} else {
			var nodeList []string
			for _, node := range resp.Node.Nodes {
				if node.Dir || !etcd.CheckInOrderKeyFormat(node.Key) {
					continue
				}
				nodeList = append(nodeList, node.Value)
			}

			nodeList = utility.RemoveDuplicateInOrder(nodeList)
			env.logger.Println("Get election qurorum size, after remove duplicates:", len(nodeList), "members:", nodeList)

			if len(nodeList) >= qurorumSize {
				membersSyncLock.Lock()
				membersElected = nodeList[:qurorumSize]
				membersSyncLock.Unlock()

				env.logger.Println("Get election qurorum finished:", membersElected)

				// Call script
				if env.eventOnReachQurorumNum != "" {
					callCmd := getCallCmdInstance("OnReachQurorumNum: ", env.eventOnReachQurorumNum)
					cmdCallWaitProcess(callCmd)
				}
				break
			}
		}
	}
}

func bootstrapLocalClusterMember() {
	//flush last log info
	defer env.logger.Sync()

	env.logger.Println("Started to boot Local cluster member, local ip:", env.localIP.String())

	if !utility.InSlice(membersElected, env.localIP.String()) {
		env.logger.Fatalln("BootstrapLocalClusterMember error, localip is not in the elected list.")
	}

	// Call script
	if env.eventOnPreStart != "" {
		callCmd := getCallCmdInstance("OnPreStart: ", env.eventOnPreStart)
		cmdCallWaitProcess(callCmd)
	}

	// Call script
	callCmd := getCallCmdInstance("OnStart: ", env.eventOnStart)
	err := callCmd.Start()
	if err != nil {
		env.logger.Println("callCmd.Start() error found:", err)
	}

	// block until cluster is up
	// no need wait group, need to termiate
	go callCmd.Wait()
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

func cmdCallWaitProcess(callCmd *exec.Cmd) {
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

	result = false
	timeCh := make(chan bool)
	go func() {
		time.Sleep(timeout)
		timeCh <- true
	}()

	// Important!!! check upstarted
	env.logger.Println("Call hook script for check discovery cluster's startup...")
	// Call script
	callCmd := getCallCmdInstance("OnPostStart: ", env.eventOnPostStart)
	cmdWaitGroup.Add(1)
	sucCh := make(chan bool)
	go func() {
		defer cmdWaitGroup.Done()

		for {
			err := callCmd.Start()
			if err != nil {
				env.logger.Println("callCmd.Start() error found:", err)
			}
			callCmd.Wait()

			env.logger.Println("Cluster is checked up now, The status is normal.")
			clusterUpdated = true

			// Check cluster if up if return exit code is normal
			if callCmd.ProcessState.Success() {
				sucCh <- true
				break
			}
		}
	}()

	select {
	case <-timeCh:
		env.logger.Println("Cluster booting is timeout, will give up and terminate.")
		exitApp.Store(true)
	case <-sucCh:
		break
	}

	return result, err
}

func watchDogRunning() {
	// Call script
	if env.eventOnClusterBooted != "" {
		callCmd := getCallCmdInstance("OnClusterBooted: ", env.eventOnClusterBooted)
		cmdCallWaitProcess(callCmd)
	}

	for {
		if isExit, ok := exitApp.Load().(bool); ok && isExit {
			break
		}

		callCmd := getCallCmdInstance("OnHealthCheck: ", env.eventOnHealthCheck)
		cmdCallWaitProcess(callCmd)
		if callCmd.ProcessState.Success() {
			// when the health check call normal return, break the infinite loop
			break
		}
	}
}

// Need to watch config size
func watchQurorumSize() {
	//flush last log info
	defer env.logger.Sync()

	kvApi := getClientKeysApi()

	for {
		watch := kvApi.Conn().Watcher(env.discoveryPath+CLUSTER_CONFIG_DIR+"/size", &client.WatcherOptions{AfterIndex: qurorumWatchIndex})
		resp, err := watch.Next(etcd.Context())

		// if Cluster is booted, quit.
		if clusterUpdated {
			break
		}

		if err != nil {
			env.logger.Fatalln("Etcd.Api() watch /config/size error:", err)
		} else {
			env.logger.Println("Etcd.Api() watch /config/size found change:", resp)

			if (resp.Action == "expire" || resp.Action == "delete") && resp.Node.Key == env.discoveryPath {
				env.logger.Println("Etcd.Api() service boot timeout reach, will delete " + env.discoveryPath + " and terminate app.")

				resp, err := kvApi.Conn().Delete(etcd.Context(), env.discoveryPath, &client.DeleteOptions{Recursive: true, Dir: true})

				// Need exit
				exitApp.Store(true)

				if err != nil {
					env.logger.Println("Etcd.Api() error while delete "+env.discoveryPath+": ", err)
				} else {
					env.logger.Println("Etcd.Api() deleted "+env.discoveryPath+" and bye: ", resp)
				}
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
	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR+"/size", &client.GetOptions{})
	if err != nil {
		env.logger.Fatalln("Etcd.Api() get /config/size error:", err)
	} else {
		env.logger.Println("Etcd.Api() get /config/size ok, Qurorum size:", resp.Node.Value)
		tmp, err := strconv.ParseInt(resp.Node.Value, 10, 64)
		if err != nil {
			env.logger.Fatalln("Error: strconv.ParseInt(/config/size) error:", err)
		}

		qurorumSyncLock.Lock()
		qurorumSize = int(tmp)
		qurorumWatchIndex = resp.Node.ModifiedIndex
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
