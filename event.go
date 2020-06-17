// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import (
	"context"
	"sync"
)

// Bus represents an event bus.
type Bus struct {
	lock sync.RWMutex
	subs map[string][]*handler
}

func (e *Bus) setup() {
	if e.subs == nil {
		e.subs = make(map[string][]*handler, 8)
	}
}

// Notify notifies listeners of an event that happened
func (e *Bus) Notify(event string, value interface{}) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	e.setup()
	if handlers, ok := e.subs[event]; ok {
		for _, h := range handlers {
			h.buffer <- value
		}
	}
}

// On registers an event listener on a system
func (e *Bus) On(event string, callback func(interface{})) context.CancelFunc {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Create the handler
	e.setup()
	ctx, cancel := context.WithCancel(context.Background())
	subscriber := &handler{
		buffer:   make(chan interface{}, 1),
		callback: &callback,
		cancel:   cancel,
	}

	// Add the listener
	e.subs[event] = append(e.subs[event], subscriber)
	go subscriber.listen(ctx)

	return e.unsubscribe(event, &callback)
}

// unsubscribe deregisters an event listener from a system
func (e *Bus) unsubscribe(event string, callback *func(interface{})) context.CancelFunc {
	return func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		if handlers, ok := e.subs[event]; ok {
			clean := make([]*handler, 0, len(handlers))
			for _, h := range handlers {
				if h.callback != callback { // Compare address
					clean = append(clean, h)
				} else {
					h.cancel()
				}
			}
		}
	}
}

// -------------------------------------------------------------------------------------------

type handler struct {
	buffer   chan interface{}
	callback *func(interface{})
	cancel   context.CancelFunc
}

// Listen listens on the buffer and invokes the callback
func (h *handler) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case value := <-h.buffer:
			(*h.callback)(value)
		}
	}
}
