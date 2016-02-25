package main

import (
	"log"
	"os"
	"os/exec"

	"zooinit/utility"
)

func main() {
	disExecCmd := "consul agent -server -data-dir=/tmp/consul -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108"

	log.Println("Etcd Discovery ExecCmd: ", disExecCmd)

	// Boot internal discovery service
	// Need to rm -rf /tmp/etcd/ because dir may be used before
	path, args, err := utility.ParseCmdStringWithParams(disExecCmd)
	if err != nil {
		log.Fatalln("Error ParseCmdStringWithParams cluster bootstrap: ", err)
	}

	clusterCmd := exec.Command(path, args...)
	clusterCmd.Stdout = os.Stdout
	clusterCmd.Stderr = os.Stdin

	err = clusterCmd.Start()
	defer clusterCmd.Process.Kill()

	clusterCmd.Wait()
}
