package utility

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	SIGNAL_HANDLE_EXIT   = true
	SIGNAL_HANDLE_NOEXIT = false
)

var (
	defaultSignalCatchSet = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGSEGV}
)

type SignalHandler func(sig os.Signal, arg interface{})
type SignalCallback struct {
	handle SignalHandler
	data   interface{}
}

// Singal catch
// 限制, 系统实现单个信号只能注册一个信号
type SignalCatcher struct {
	handles map[os.Signal][]*SignalCallback

	// WaitGroup for goroutine to complete
	wg sync.WaitGroup

	// Is exit after signal process
	exitAfterCall bool
}

func NewSignalCatcher() *SignalCatcher {
	handles := make(map[os.Signal][]*SignalCallback)
	return &SignalCatcher{handles: handles, exitAfterCall: false}
}

func NewSignalCallback(handle SignalHandler, data interface{}) *SignalCallback {
	return &SignalCallback{handle: handle, data: data}
}

func (s *SignalCatcher) SetDefault(handle *SignalCallback) bool {
	if s == nil {
		return false
	}

	return s.SetHandle(defaultSignalCatchSet, handle)
}

func (s *SignalCatcher) SetHandle(sigs []os.Signal, handle *SignalCallback) bool {
	if s == nil {
		return false
	}

	for _, sig := range sigs {
		s.handles[sig] = append(s.handles[sig], handle)
	}

	return true
}

func (s *SignalCatcher) GetSignalList() []os.Signal {
	if s == nil {
		return nil
	}

	var list []os.Signal
	for sig, _ := range s.handles {
		list = append(list, sig)
	}

	return list
}

func (s *SignalCatcher) GetSignalStringList() []string {
	if s == nil {
		return nil
	}

	var list []string
	for sig, _ := range s.handles {
		list = append(list, sig.String())
	}

	return list
}

// internal signal process handle
func (s *SignalCatcher) handle(sig os.Signal) {
	if s == nil {
		return
	}

	if calllist, ok := s.handles[sig]; ok {
		for _, call := range calllist {
			//wait group
			s.wg.Add(1)

			go func() {
				defer s.wg.Done()
				call.handle(sig, call.data)
			}()
		}
	}
}

// support multiple signal
func (s *SignalCatcher) RegisterAndServe() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, s.GetSignalList()...)

	go func() {
		for {
			sig := <-c
			//print("receive sig: ", sig.String(), "\n")
			s.handle(sig)

			s.wg.Wait()

			if s.exitAfterCall {
				os.Exit(1)
			}
		}
	}()
}

// whether IsExitEnable
func (s *SignalCatcher) IsExit() bool {
	if s == nil {
		return false
	}

	return s.exitAfterCall
}

func (s *SignalCatcher) EnableExit() {
	if s == nil {
		return
	}

	s.exitAfterCall = SIGNAL_HANDLE_EXIT
}

func (s *SignalCatcher) DisableExit() {
	if s == nil {
		return
	}

	s.exitAfterCall = SIGNAL_HANDLE_NOEXIT
}
