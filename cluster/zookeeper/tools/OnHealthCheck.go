package main

import (
	"fmt"
	"os"

	"github.com/samuel/go-zookeeper/zk"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("OnHealthCheck error params: OnHealthCheck clientip")
		os.Exit(1)
	}

	clientip := os.Args[1]
	zkc, _, err := zk.Connect([]string{clientip}, time.Second)
	if err != nil {
		fmt.Println("Connect returned error:", err)
		os.Exit(1)
	}

	tryPath := "/zookeeper"
	_, stat, err := zkc.Get(tryPath)
	if err != nil {
		fmt.Println("Get returned error:", err)
		os.Exit(1)
	} else if stat == nil {
		fmt.Println("Get returned nil stat")
		os.Exit(1)
	}

	fmt.Println("Fetch Stat data:", stat.Czxid)

	defer zkc.Close()
}
