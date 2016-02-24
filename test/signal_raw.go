package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"strings"
	"zooinit/utility"
)

// kill -9是不可捕捉的，应该直接kill或者ctrl+c
// callback signature must be comptable
func cleanup(sig os.Signal, data interface{}) {
	fmt.Println("Catched signal, cleanup. sig: ", sig, "args: ", data)
}

type Hello int

func main() {
	fmt.Println("Started to test SignalTest")
	var a Hello = 1
	fmt.Println("Local type a:", a)

	clean := func(sig os.Signal, data interface{}) {
		fmt.Println("Catched signal, cleanup. sig: ", sig, "string: ", data)
	}

	sg := utility.NewSignalCatcher()
	call := utility.NewSignalCallback(clean, "this is test cleanup")
	fmt.Println("catch list init:", strings.Join(sg.GetSignalStringList(), ", "))

	sg.SetDefault(call)
	fmt.Println("catch list after SetDefault:", strings.Join(sg.GetSignalStringList(), ", "))

	usercall := utility.NewSignalCallback(cleanup, "this is test cleanup")
	sg.SetHandle([]os.Signal{syscall.SIGUSR1, syscall.SIGUSR2}, usercall)

	fmt.Println("catch list:", strings.Join(sg.GetSignalStringList(), ", "))

	fmt.Println("You can send signal in 10s..")
	sg.RegisterAndServe()

	time.Sleep(10 * time.Second)
	fmt.Println("Wait completed.")
}
