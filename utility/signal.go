package utility

import (
	"errors"
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
	handler SignalHandler
	data    interface{}
}

// Singal catch
// 限制, 系统实现单个信号只能注册一个信号
// 可以通过signal_callback实现单个信号执行多个callback
type SignalCatcher struct {
	handlers map[os.Signal]*SignalCallStack

	// WaitGroup for goroutine to complete
	wg *sync.WaitGroup

	// Is exit after signal process
	exitAfterCall bool
}

func NewSignalCatcher() *SignalCatcher {
	handlers := make(map[os.Signal]*SignalCallStack)
	wg := &sync.WaitGroup{}
	return &SignalCatcher{handlers: handlers, exitAfterCall: false, wg: wg}
}

func NewSignalCallback(handler SignalHandler, data interface{}) *SignalCallback {
	return &SignalCallback{handler: handler, data: data}
}

func (s *SignalCatcher) SetDefault(handler *SignalCallStack) (bool, error) {
	return s.SetHandler(defaultSignalCatchSet, handler)
}

func (s *SignalCatcher) SetHandler(sigs []os.Signal, handler *SignalCallStack) (bool, error) {
	for _, sig := range sigs {
		if _, ok := s.handlers[sig]; ok {
			return false, errors.New("SignalCallStack of sig:" + sig.String() + " is set, please refer SignalCallStack to manipulate.")
		}
	}

	for _, sig := range sigs {
		s.handlers[sig] = handler
	}

	return true, nil
}

func (s *SignalCatcher) GetHandler(sig os.Signal) (*SignalCallStack, error) {
	handler, ok := s.handlers[sig]
	if !ok {
		return nil, errors.New("SignalCallStack of sig:" + sig.String() + " is NOT set.")
	}

	return handler, nil
}

func (s *SignalCatcher) GetSignalList() []os.Signal {
	var list []os.Signal
	for sig, _ := range s.handlers {
		list = append(list, sig)
	}

	return list
}

func (s *SignalCatcher) GetSignalStringList() []string {
	var list []string
	for sig, _ := range s.handlers {
		list = append(list, sig.String())
	}

	return list
}

// internal signal process handle
func (s *SignalCatcher) handle(sig os.Signal) {
	if callStack, ok := s.handlers[sig]; ok {
		callStack.Trigger(sig, s.wg)
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
	return s.exitAfterCall
}

func (s *SignalCatcher) EnableExit() {
	s.exitAfterCall = SIGNAL_HANDLE_EXIT
}

func (s *SignalCatcher) DisableExit() {
	s.exitAfterCall = SIGNAL_HANDLE_NOEXIT
}
