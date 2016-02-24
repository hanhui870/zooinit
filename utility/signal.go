package utility

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
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
type SignalCatcher struct {
	handles map[os.Signal][]*SignalCallback
	// WaitGroup for goroutine to complete
	wg sync.WaitGroup
}

func NewSignalCatcher() *SignalCatcher {
	handles := make(map[os.Signal][]*SignalCallback)
	return &SignalCatcher{handles: handles}
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

func (s *SignalCatcher) RegisterAndServe() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, s.GetSignalList()...)

	go func() {
		sig := <-c
		//print("receive sig: ", sig.String(), "\n")
		s.handle(sig)

		s.wg.Wait()
		os.Exit(1)
	}()
}
