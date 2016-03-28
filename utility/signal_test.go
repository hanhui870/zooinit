package utility

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

// kill -9是不可捕捉的，应该直接kill或者ctrl+c
// callback signature must be comptable
func cleanup(sig os.Signal, data interface{}) {
	fmt.Println("Catched signal, cleanup. sig: ", sig, "args: ", data)
}

type Hello int

func TestSignalTest(t *testing.T) {
	t.Log("Started to test SignalTest")
	var a Hello = 1
	t.Log("Local type a:", a)

	clean := func(sig os.Signal, data interface{}) {
		fmt.Println("Catched signal, cleanup. sig: ", sig, "string: ", data)
	}

	sg := NewSignalCatcher()
	call := NewSignalCallback(clean, "this is test DEFAULT cleanup")
	stack := NewSignalCallStack()
	sg.SetDefault(stack)
	stack.Add(call)

	// The same handler will affect all signal.
	stackFetch, err := sg.GetHandler(syscall.SIGINT)
	if err != nil {
		t.Error("Found error. Fetch sg.GetHandler(syscall.SIGINT) failed, ", err)
	} else {
		stackFetch.Add(NewSignalCallback(clean, "this is GetHandler cleanup"))
	}

	usercall := NewSignalCallback(cleanup, "this is test User cleanup")
	_, err = sg.SetHandler([]os.Signal{syscall.SIGUSR1}, NewSignalCallStack())
	if err != nil {
		t.Error("Found error. Fetch sg.GetHandler(syscall.SIGUSR1) failed, ", err)
	}

	stackFetch2, err := sg.GetHandler(syscall.SIGUSR1)
	if err != nil {
		t.Error("Found error. Fetch sg.GetHandler(syscall.SIGUSR1) failed, ", err)
	} else {
		stackFetch2.Add(usercall)
	}

	t.Log("catch list:", sg.GetSignalList())

	t.Log("You can send signal in 10s..")
	sg.RegisterAndServe()

	time.Sleep(10 * time.Second)
	t.Log("Wait completed.")
}
