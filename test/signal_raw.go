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
	fmt.Println("catch list init:", strings.Join(sg.GetSignalStringList(), ", "))

	//call := utility.NewSignalCallback(clean, "this is test cleanup")
	//sg.SetDefault(call)

	usercall := utility.NewSignalCallback(cleanup, "this is user signal, and will exit...")
	sg.SetHandle([]os.Signal{syscall.SIGUSR1, syscall.SIGUSR2}, usercall)

	fmt.Println("catch list:", strings.Join(sg.GetSignalStringList(), ", "))

	tw := 100 * time.Second
	fmt.Println("You can send signal in ", tw)
	//support exit need to exit kill -30 pid
	sg.EnableExit()
	sg.RegisterAndServe()

	// Second signal book
	sg2 := utility.NewSignalCatcher()
	call2 := utility.NewSignalCallback(clean, "this is test cleanup call2")
	fmt.Println("catch list init2:", strings.Join(sg2.GetSignalStringList(), ", "))

	sg2.SetDefault(call2)
	fmt.Println("catch list after SetDefault:", strings.Join(sg2.GetSignalStringList(), ", "))
	sg2.RegisterAndServe()

	time.Sleep(tw)
	fmt.Println("Wait completed.")
}
