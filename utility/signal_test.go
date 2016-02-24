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
	call := NewSignalCallback(clean, "this is test cleanup")
	sg.SetDefault(call)
	usercall := NewSignalCallback(cleanup, "this is test cleanup")
	sg.SetHandle([]os.Signal{syscall.SIGUSR1}, usercall)

	t.Log("catch list:", sg.GetSignalList())

	t.Log("You can send signal in 10s..")
	sg.RegisterAndServe()

	time.Sleep(10 * time.Second)
	t.Log("Wait completed.")
}
