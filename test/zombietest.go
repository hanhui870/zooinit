package main

import (
	"log"
	"os/exec"

	"os"
	"time"
)

func main() {
	callCmd := exec.Command("etcd", "--data-dir", "/tmp/etcd")
	callCmd.Stderr = os.Stderr
	callCmd.Stdout = os.Stdout
	callCmd.Stdin = os.Stdin

	err := callCmd.Start()
	if err != nil {
		log.Println("callCmd.Start() error found:", err)
	}
	// 没有callCmd.Wait(), etcd被杀后, etcd肯定是僵尸进程
	// 有GO wait之后,正常结束了进程,没有僵尸经常.
	go callCmd.Wait()

	time.Sleep(time.Hour)
}
