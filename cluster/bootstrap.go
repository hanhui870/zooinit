package cluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	syslog "log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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
	// booted uuid of servers
	CLUSTER_CONFIG_DIR_BOOTED_UUIDS = CLUSTER_CONFIG_DIR + "/booted_uuids"
	// map uuid of servers ip: uuid->ip
	CLUSTER_CONFIG_DIR_BOOTED_UUID_MAP = CLUSTER_CONFIG_DIR + "/booted_uuid_map"
	CLUSTER_ELECTION_DIR               = "/election"
	// 6. check health update this
	CLUSTER_MEMBER_DIR = "/members"
	// member node ttl
	CLUSTER_MEMBER_NODE_TTL = 1 * time.Minute

	// service name reserved for bootstrap
	BOOTSTRAP_SERVICE_NAME = "bootstrap"

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
	membersElected []string
	uuidElected    []string
	// map uuid=>ip
	uuidMap         map[string]string
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
	uuidMap = make(map[string]string, 3)
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
	} else if service == BOOTSTRAP_SERVICE_NAME {
		syslog.Fatalln("Service name of \"" + BOOTSTRAP_SERVICE_NAME + "\" is reserved.")
	}

	// backend of servie
	backend := c.String("backend")
	if backend == "" {
		backend = service
	}

	env = NewEnvInfo(iniobj, backend, service, c)

	env.GuaranteeSingleRun()

	//flush last log info
	defer env.Logger.Sync()
	//register signal watcher
	env.RegisterSignalWatch()

	env.Logger.Println("Server UUID:", env.UUID)
	env.Logger.Println("Logger channel:", env.LogChannel)
	if env.LogChannel != log.LOG_STDOUT {
		env.Logger.Println("Logger path:", env.LogPath)
	}
	env.Logger.Println("Timeout:", env.Timeout.String())
	env.Logger.Println("Qurorum:", env.Qurorum)
	env.Logger.Println("Discover method:", env.discoveryMethod)
	env.Logger.Println("Discover path:", env.discoveryPath)
	env.Logger.Println("Health check interval:", env.HealthCheckInterval)
	env.Logger.Println("env.discoveryTarget for fetch members:", env.discoveryTarget)

	// update endpoints
	UpdateLatestEndpoints()

	// register node
	initializeClusterDiscoveryInfo()

	// loop wait qurorum size of nodes is registed
	loopUntilQurorumIsReached()

	// start up local node
	bootstrapLocalClusterMember()

	//must before watchDogRunning, can before cluster is up.
	go clusterMemberRestartRoutine()

	// Will block
	loopUntilClusterIsUp(env.Timeout)

	// Apart from watch dog running
	eventOnClusterBooted()

	// watch and check cluster health [watchdog], block until server receive term signal
	watchDogRunning()

	// final wait.
	cmdWaitGroup.Wait()
	env.Logger.Println("App runtime reaches end.")
}

// Fetch bootstrap env instance
func GetEnvInfo() *envInfo {
	return env
}

// init cluster bootstrap info
func initializeClusterDiscoveryInfo() {
	//flush last log info
	defer env.Logger.Sync()

	kvApi := getClientKeysApi()

	// Set qurorum size
	resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath, "", &client.SetOptions{Dir: true, TTL: env.Timeout, PrevExist: client.PrevNoExist})
	if err != nil {
		//ignore exist error
		if !etcd.EqualEtcdError(err, client.ErrorCodeNodeExist) {
			// check if exist need to add qurorum
			env.Logger.Fatalln("Etcd.Api() set "+env.discoveryPath+" error:", err)
		} else {
			env.Logger.Println("Etcd.Api() set " + env.discoveryPath + " notice: node exist, will add qurorum directly.")

			resp, err = kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED, &client.GetOptions{})
			if err == nil {
				env.Logger.Println("Etcd.Api() cluster already booted at :", CLUSTER_CONFIG_DIR_BOOTED, ", will check whether restart needed.")

				// fetch latest node list
				// Can not use ip for check, Docker restart may change ip.
				if IsClusterBootedBefore() {
					env.Logger.Println("Zooinit has found cluster has started before, will continue to restart...")

					//remove /election out dated item compare to /members
					removeOutDateClusterMemberElectionAndUUIDCheck()

					clusterIsBootedBefore = true
				} else {
					env.Logger.Fatalln("Zooinit found cluster has NOT started before, it is not allowed to join a booted cluster.")
				}

			} else if !etcd.EqualEtcdError(err, client.ErrorCodeKeyNotFound) {
				env.Logger.Fatalln("Etcd.Api() found error while fetch:", CLUSTER_CONFIG_DIR_BOOTED, " error:", err)
			}
		}

	} else {
		// Create success, set config
		env.Logger.Println("Etcd.Api() set "+env.discoveryPath+" ok, TTL:", env.Timeout.String(), resp)

		// Set config size
		resp, err = kvApi.Conn().Set(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_SIZE, strconv.Itoa(env.Qurorum), &client.SetOptions{PrevExist: client.PrevNoExist})
		if err != nil {
			env.Logger.Fatalln("Etcd.Api() set "+CLUSTER_CONFIG_DIR_SIZE+" error:", err)
		} else {
			env.Logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_SIZE+" ok, Qurorum size:", env.Qurorum, resp)
		}
	}

	// Call script
	if env.eventOnPreRegist != "" {
		callCmd := getCallCmdInstance("OnPreRegist: ", env.eventOnPreRegist)
		cmdCallWaitProcessSync(callCmd)
	}

	// Create qurorum in order node
	election := NewElectionMemberEnv(env)
	resp, err = kvApi.Conn().CreateInOrder(etcd.Context(), env.discoveryPath+CLUSTER_ELECTION_DIR, election.ToJson(), nil)
	if err != nil {
		env.Logger.Fatalln("Etcd.Api() CreateInOrder error:", err)
	} else {
		env.Logger.Println("Etcd.Api() CreateInOrder ok:", resp)

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
	defer env.Logger.Sync()

	memApi := getClientMembersApi(strings.Split(env.discoveryTarget, ","))

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

func loopUntilQurorumIsReached() {
	//flush last log info
	defer env.Logger.Sync()

	kvApi := getClientKeysApi()

	// GetLatestElectionMember index incrs from this one
	if !clusterIsBootedBefore {
		resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_ELECTION_DIR, &client.GetOptions{Recursive: true, Sort: true})
		if err != nil {
			env.Logger.Fatalln("Etcd.Api() get "+env.discoveryPath+CLUSTER_ELECTION_DIR+" lastest ModifiedIndex error:", err)

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
		wather := kvApi.Conn().Watcher(env.discoveryPath+CLUSTER_ELECTION_DIR, &client.WatcherOptions{Recursive: true, AfterIndex: latestIndex})
		resp, err := wather.Next(etcd.Context())
		if err != nil {
			env.Logger.Fatalln("Etcd.Watcher() watch "+env.discoveryPath+CLUSTER_ELECTION_DIR+" error:", err)

		} else {
			latestIndex = resp.Node.ModifiedIndex
			env.Logger.Println("Get last ModifiedIndex of watch:", latestIndex)
		}

		// fetch latest node list
		nodeList, uuidList, err := getLastestNodeIPList()
		if err != nil {
			env.Logger.Println("Etcd.Api() get "+env.discoveryPath+CLUSTER_ELECTION_DIR+" lastest election nodes error:", err)
			env.Logger.Println("Will exit now...")

			//Need to expecility exit
			//Sleep for finish action at watchQurorumSize()
			for {
				if isExit, ok := exitApp.Load().(bool); ok && isExit {
					env.Logger.Println("Wait completed, bye...")
					os.Exit(0)
				}

				time.Sleep(100 * time.Millisecond)
			}

		} else {
			env.Logger.Println("Get election qurorum size, after remove duplicates:", len(nodeList), "members:", nodeList)

			if len(nodeList) >= qurorumSize {
				membersSyncLock.Lock()
				membersElected = nodeList[:qurorumSize]
				uuidElected = uuidList[:qurorumSize]
				membersSyncLock.Unlock()

				// map uuid=>ip
				for iter, ip := range membersElected {
					uuidMap[uuidElected[iter]] = ip
				}

				env.Logger.Println("Get election qurorum finished:, IP:", membersElected, " UUID:", uuidElected)

				setClusterBootedUUIDs(uuidElected)

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

func bootstrapLocalClusterMember() {
	//flush last log info
	defer env.Logger.Sync()

	env.Logger.Println("Started to boot Local cluster member, local ip:", env.LocalIP.String())

	if !utility.InSlice(membersElected, env.LocalIP.String()) {
		env.Logger.Fatalln("BootstrapLocalClusterMember error, localip is not in the elected list.")
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
		env.Logger.Println("callCmd.Start() error found:", err)
	}

	// block until cluster is up
	cmdWaitGroup.Add(1)
	go func() {
		defer cmdWaitGroup.Done()
		err = callCmdStartInstance.Wait()
		if err != nil {
			env.Logger.Println("callCmd.Wait() finished with error found:", err)
		} else {
			env.Logger.Println("callCmd.Wait() finished without error.")
		}

		if isExit, ok := exitApp.Load().(bool); !ok || !isExit {
			env.Logger.Println("BootstrapLocalClusterMember do not detect exitApp cmd, will restart...")
			restartMemberChannel <- MEMBER_RESTART_CMDWAIT
		}
	}()
}

func getCallCmdInstance(logPrefix, event string) *exec.Cmd {
	callCmd := exec.Command("bash", "-c", "script/entrypoint.sh")

	loggerIOAdapter := log.NewLoggerIOAdapter(env.Logger)
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
			env.Logger.Println("callCmd.Start() error found:", err)
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
		env.Logger.Println("callCmd.Start() error found:", err)
	}
	callCmd.Wait()
}

func getCallCmdENVSet(event string) []string {

	envs := []string{"ZOOINIT_CLUSTER_BACKEND=" + env.clusterBackend}
	envs = append(envs, "ZOOINIT_CLUSTER_SERVICE="+env.Service)
	envs = append(envs, "ZOOINIT_CLUSTER_EVENT="+event)
	envs = append(envs, "ZOOINIT_SERVER_IP_LIST="+strings.Join(membersElected, ","))
	envs = append(envs, "ZOOINIT_LOCAL_IP="+env.LocalIP.String())

	// Master first one
	if len(membersElected) > 0 {
		envs = append(envs, "ZOOINIT_MASTER_IP="+membersElected[0])
	} else {
		envs = append(envs, "ZOOINIT_MASTER_IP=")
	}

	if len(uuidMap) > 0 {
		buf, err := json.Marshal(uuidMap)
		if err == nil {
			envs = append(envs, "ZOOINIT_SERVER_UUID_MAP="+string(buf))
		} else {
			env.Logger.Fatal("getCallCmdENVSet: json.Marshal(uuidMap) error:", err)
		}
	} else {
		envs = append(envs, "ZOOINIT_SERVER_UUID_MAP=")
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
	defer env.Logger.Sync()

	kvApi := getClientKeysApi()

	result = false

	// Important!!! check upstarted
	env.Logger.Println("Call hook script for check discovery cluster's startup...")
	// Call script
	cmdWaitGroup.Add(1)
	sucCh := make(chan bool)
	go func() {
		defer cmdWaitGroup.Done()

		failedTimes := 0
		for {
			//Need init every call: callCmd.Start() error found: exec: already started
			callCmd := getCallCmdInstance("OnPostStart: ", env.eventOnPostStart)

			err := callCmd.Start()
			if err != nil {
				env.Logger.Println("callCmd.Start() error found:", err)
			}
			callCmd.Wait()

			// Check cluster if up if return exit code is normal
			if callCmd.ProcessState.Success() {
				sucCh <- true

				env.Logger.Println("Cluster is checked up now, The status is normal.")
				clusterUpdated = true

				//mark cluster booted
				makeClusterBooted()

				// schedule to update discovery path ttl
				go updateDiscoveryTTL()

				// Set config/booted to true
				booted := "true," + env.LocalIP.String() + "," + time.Now().String()
				resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED, booted, &client.SetOptions{PrevExist: client.PrevNoExist})
				if err != nil {
					//ignore exist error
					if etcd.EqualEtcdError(err, client.ErrorCodeNodeExist) {
						// check if exist need to add qurorum
						env.Logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED+" set by another node, error:", err)

						resp, err = kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_BOOTED, &client.GetOptions{})
						if err == nil {
							env.Logger.Println("Etcd.Api() cluster already booted at :", CLUSTER_CONFIG_DIR_BOOTED, "Resp:", resp)
						} else {
							env.Logger.Fatalln("Etcd.Api() found error while fetch:", CLUSTER_CONFIG_DIR_BOOTED, " error:", err)
						}
					} else {
						env.Logger.Fatalln("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED+" error:", err)
					}
				} else {
					env.Logger.Println("Etcd.Api() set "+CLUSTER_CONFIG_DIR_BOOTED+" ok:", booted, "Resp:", resp)
				}
				break
			}

			failedTimes++
			tts := time.Duration(failedTimes) * time.Second
			env.Logger.Println("OnPostStart failed " + strconv.Itoa(failedTimes) + ", will sleep " + tts.String() + " continue to call event...")
			time.Sleep(tts)
		}
	}()

	select {
	case <-time.After(env.Timeout):
		env.Logger.Println("Cluster booting is timeout, will give up and terminate.")
		exitApp.Store(true)
	case <-sucCh:
		break
	}

	return result, err
}

func getBootedFlagFile() string {
	return env.LogPath + "/booted"
}

func makeClusterBooted() (bool, error) {
	//flush last log info
	defer env.Logger.Sync()

	//create if nessary
	err := os.MkdirAll(filepath.Dir(getBootedFlagFile()), log.DEFAULT_LOGDIR_MODE)
	if err != nil {
		return false, &os.PathError{"create dir", getBootedFlagFile(), err}
	}

	file, err := os.OpenFile(getBootedFlagFile(), os.O_CREATE|os.O_TRUNC|os.O_RDWR, log.DEFAULT_LOGFILE_MODE)
	if err != nil {
		return false, &os.PathError{"open", getBootedFlagFile(), err}
	}

	defer file.Close()
	n, err := file.Write(bytes.NewBufferString(time.Now().Format(time.RFC3339)).Bytes())
	if err != nil {
		return false, err
	} else if n == 0 {
		err = errors.New("Write 0 length content for makeClusterBooted()")
		return false, err
	}

	return true, nil
}

func IsClusterBootedBefore() bool {
	//flush last log info
	defer env.Logger.Sync()

	file, err := os.Open(getBootedFlagFile())
	if err == nil {
		content, err := ioutil.ReadAll(file)
		if err == nil {
			ti, err := time.Parse(time.RFC3339, string(content))
			if err == nil {
				env.Logger.Println("Cluster ever booted at:", ti.Format(time.RFC3339))
				return true
			}
		}
	}

	return false
}

// usage: go updateDiscoveryTTL()
// update need to call after
func updateDiscoveryTTL() {
	//flush last log info
	defer env.Logger.Sync()
	rand.Seed(time.Now().UnixNano())

	//30-59s 随机更新一次,所有节点都会更新
	failedTimes := 0
	for {
		kvApi := getClientKeysApi()

		// Update TTL
		resp, err := kvApi.Conn().Set(etcd.Context(), env.discoveryPath, "", &client.SetOptions{Dir: true, TTL: env.Timeout, PrevExist: client.PrevExist})
		if err != nil {
			// Not exit.
			env.Logger.Println("Etcd.Api() update "+env.discoveryPath+" TTL error:", err, " faildTimes:", failedTimes)
		} else {
			env.Logger.Println("Etcd.Api() update "+env.discoveryPath+" ok", "Resp:", resp)
			failedTimes = 0
		}

		next := time.Duration(30 + rand.Intn(30))
		time.Sleep(next * time.Second)
	}
}

func eventOnClusterBooted() {
	//flush last log info
	defer env.Logger.Sync()

	// Call OnClusterBooted script, no need goroutine
	if env.eventOnClusterBooted != "" {
		callCmd := getCallCmdInstance("OnClusterBooted: ", env.eventOnClusterBooted)
		err := callCmd.Start()
		if err != nil {
			env.Logger.Println("callCmd.Start() error found:", err)
		}
		callCmd.Wait()
	}
}

// Check cluster health after cluster is up.
func watchDogRunning() {
	//flush last log info
	defer env.Logger.Sync()

	//this can not use goroutine, this is a loop
	firstRun := true
	//failedTimes update to etcd
	failedTimes := 0
	for {
		if isExit, ok := exitApp.Load().(bool); ok && isExit {
			env.Logger.Println("Receive exitApp signal, break watchDogRunning loop.")
			break
		}

		callCmd := getCallCmdInstance("OnHealthCheck: ", env.eventOnHealthCheck)
		if !firstRun {
			// like exec health check, no need to debug env info every time.
			callCmd.Env = append(callCmd.Env, "ZOOINIT_SILENT_ENV_INFO=true")
		}

		// may runtime error: invalid memory address or nil pointer dereference
		cmdCallWaitProcessSync(callCmd)

		var cm *ClusterMember
		var result bool
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
				env.Logger.Println("Cluster member is NOT healthy, will trigger Restart. Failed times:", execCheckFailedTimes, ", MEMBER_MAX_FAILED_TIMES:", MEMBER_MAX_FAILED_TIMES)
				execCheckFailedTimes = 0
				restartMemberChannel <- MEMBER_RESTART_HEALTHCHECK
			} else {
				env.Logger.Println("Cluster member is NOT healthy, Failed times:", execCheckFailedTimes)
			}
		}

		cm = NewClusterMember(env, result, execCheckFailedTimes)
		kvApi := getClientKeysApi()
		pathNode := env.discoveryPath + CLUSTER_MEMBER_DIR + "/" + env.GetNodename()
		resp, err := kvApi.Conn().Set(etcd.Context(), pathNode, cm.ToJson(), &client.SetOptions{Dir: false, TTL: CLUSTER_MEMBER_NODE_TTL})
		if err != nil {
			env.Logger.Println("Etcd.Api() update "+pathNode+" State error:", err, " faildTimes:", failedTimes)
		} else {
			env.Logger.Println("Etcd.Api() update "+pathNode+" ok", "Resp:", resp)
			failedTimes = 0
		}

		if firstRun {
			firstRun = false
		}

		// sleep interval time
		time.Sleep(env.HealthCheckInterval)
	}
}

// Need to watch config size
func watchQurorumSize() {
	//flush last log info
	defer env.Logger.Sync()

	for {
		kvApi := getClientKeysApi()
		watch := kvApi.Conn().Watcher(env.discoveryPath+CLUSTER_CONFIG_DIR_SIZE, &client.WatcherOptions{AfterIndex: qurorumWatchIndex})
		resp, err := watch.Next(etcd.Context())

		// if Cluster is booted, quit.
		if clusterUpdated {
			break
		}

		if err != nil {
			env.Logger.Fatalln("Etcd.Api() watch "+CLUSTER_CONFIG_DIR_SIZE+" error:", err)
		} else {
			env.Logger.Println("Etcd.Api() watch "+CLUSTER_CONFIG_DIR_SIZE+" found change:", resp)

			if (resp.Action == "expire" || resp.Action == "delete") && resp.Node.Key == env.discoveryPath {
				env.Logger.Println("Etcd.Api() service boot timeout reach, will delete " + env.discoveryPath + " and terminate app.")

				resp, err := kvApi.Conn().Delete(etcd.Context(), env.discoveryPath, &client.DeleteOptions{Recursive: true, Dir: true})

				if err != nil {
					env.Logger.Println("Etcd.Api() error while delete "+env.discoveryPath+": ", err)
				} else {
					env.Logger.Println("Etcd.Api() deleted "+env.discoveryPath+" and bye: ", resp)
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
	defer env.Logger.Sync()

	kvApi := getClientKeysApi()
	// get config size
	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_CONFIG_DIR_SIZE, &client.GetOptions{})
	if err != nil {
		env.Logger.Fatalln("Etcd.Api() get "+CLUSTER_CONFIG_DIR_SIZE+" error:", err)
	} else {
		env.Logger.Println("Etcd.Api() get "+CLUSTER_CONFIG_DIR_SIZE+" ok, Qurorum size:", resp.Node.Value)
		tmp, err := strconv.ParseInt(resp.Node.Value, 10, 64)
		if err != nil {
			env.Logger.Fatalln("Error: strconv.ParseInt("+CLUSTER_CONFIG_DIR_SIZE+") error:", err)
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
			if trigger == MEMBER_RESTART_CMDWAIT {
				env.Logger.Println("exec restartMemberChannel MEMBER_RESTART_CMDWAIT...")
				// need to reset execCheckFailedTimes
				execCheckFailedTimes = 0

			} else if trigger == MEMBER_RESTART_HEALTHCHECK {
				env.Logger.Println("exec restartMemberChannel MEMBER_RESTART_HEALTHCHECK...")
				if callCmdStartInstance.Process != nil {
					env.Logger.Println("Kill old process runtime, pid:", callCmdStartInstance.Process.Pid)
					callCmdStartInstance.Process.Kill()
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
			if callCmdStartInstance.ProcessState != nil {
				env.Logger.Println("Exception: callCmdStartInstance.ProcessState is:", callCmdStartInstance.ProcessState.String())
				bootstrapLocalClusterMember()

			} else {
				env.Logger.Println("Exception: callCmdStartInstance.ProcessState is nil.")
			}
		}
	}
}

// remove /election out dated item compare to /members
// can also remove self, because will register later
// 03.30 must use ip and uuid compare, for ip may change.
func removeOutDateClusterMemberElectionAndUUIDCheck() {
	//flush last log info
	defer env.Logger.Sync()

	env.Logger.Println("Process removeOutDateClusterMemberElection: remove /election out dated item compare to /members...")

	memlist, err := getLastestClusterMemberList()
	if err != nil {
		env.Logger.Fatalln("Failed to call getLastestClusterMemberList:", err)
	}
	memIpList := []string{}
	for _, mem := range memlist {
		memIpList = append(memIpList, mem.Localip)
	}
	env.Logger.Println("Fetched latest membesrs:", memIpList)

	uuidList, err := getClusterBootedUUIDs()
	//reboot must have uuidList
	if err != nil {
		env.Logger.Fatalln("Failed to call getClusterBootedUUIDs:", err)
	}
	if !utility.InSlice(uuidList, env.UUID) {
		env.Logger.Fatalln("Endpoint uuid " + env.UUID + " is not in the booted uuidList, will give up reboot and terminate.")
	}

	kvApi := getClientKeysApi()
	resp, err := kvApi.Conn().Get(etcd.Context(), env.discoveryPath+CLUSTER_ELECTION_DIR, &client.GetOptions{Recursive: true, Sort: true})
	if err == nil {
		for _, node := range resp.Node.Nodes {
			if node.Dir || !etcd.CheckInOrderKeyFormat(node.Key) {
				env.Logger.Println("error CheckInOrderKeyFormat or dir, skip:", node.Key)
				continue
			}

			em, err := BuildElectionMemberFromJSON(node.Value)
			if err != nil {
				env.Logger.Println("Fetch Invaild electioin node:", node.Value, " will continue..")
				continue
			}
			//if not in the memberlist, need to clear
			if !utility.InSlice(memIpList, em.Ip) {
				env.Logger.Println("Endpoint ip "+em.Ip+" is not in the memberlist, will be deleted, key:", node.Key)
				resp, err = kvApi.Conn().Delete(etcd.Context(), node.Key, &client.DeleteOptions{Recursive: false, Dir: false})
				if err != nil {
					env.Logger.Println("RemoveOutDateClusterMemberElection error delete key:", node.Key)
				}
			}

			//if not in the uuidlist, need to clear
			if !utility.InSlice(uuidList, em.Uuid) {
				env.Logger.Println("Endpoint uuid "+em.Uuid+" is not in the uuidList, will be deleted, key:", node.Key)
				resp, err = kvApi.Conn().Delete(etcd.Context(), node.Key, &client.DeleteOptions{Recursive: false, Dir: false})
				if err != nil {
					env.Logger.Println("RemoveOutDateClusterMemberElection error delete key:", node.Key)
				}
			}
		}
	}
}
