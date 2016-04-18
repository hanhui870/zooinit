// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package utility

import (
	"os"
	"sync"
)

// call stack may change after registered
type SignalCallStack struct {
	handlers []*SignalCallback
}

func NewSignalCallStack() *SignalCallStack {
	return &SignalCallStack{handlers: make([]*SignalCallback, 10)}
}

func (s *SignalCallStack) Add(handler *SignalCallback) {
	s.handlers = append(s.handlers, handler)
}

// Delete handle from repo
func (s *SignalCallStack) Delete(handler *SignalCallback) (result bool) {
	for key, value := range s.handlers {
		if value == handler {
			s.handlers[key] = nil
			result = true
		}
	}

	return result
}

// Trigger action like defer, last in first trigger, NOT gurantee
func (s *SignalCallStack) Trigger(sig os.Signal, wg *sync.WaitGroup) {
	lenCall := len(s.handlers)

	// index start from 0
	for iter := lenCall - 1; iter >= 0; iter-- {
		handler := s.handlers[iter]
		if handler != nil {
			//wait group
			wg.Add(1)
			go func() {
				defer wg.Done()
				handler.handler(sig, handler.data)
			}()
		}
	}
}
