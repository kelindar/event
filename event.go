// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Event represents an event contract
type Event interface {
	Type() uint32
}

// registry holds an immutable sorted array of event mappings
type registry struct {
	keys []uint32 // Event types (sorted)
	subs []any    // Corresponding subscribers
}

// ------------------------------------- Dispatcher -------------------------------------

// Dispatcher represents an event dispatcher.
type Dispatcher struct {
	subs atomic.Pointer[registry] // Atomic pointer to immutable array
	done chan struct{}            // Cancellation
	df   time.Duration            // Flush interval
	mu   sync.Mutex               // Only for writes (subscribe/unsubscribe)
}

// NewDispatcher creates a new dispatcher of events.
func NewDispatcher() *Dispatcher {
	d := &Dispatcher{
		df:   500 * time.Microsecond,
		done: make(chan struct{}),
	}

	d.subs.Store(&registry{
		keys: make([]uint32, 0, 16),
		subs: make([]any, 0, 16),
	})
	return d
}

// Close closes the dispatcher
func (d *Dispatcher) Close() error {
	close(d.done)
	return nil
}

// isClosed returns whether the dispatcher is closed or not
func (d *Dispatcher) isClosed() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
}

// findGroup performs a lock-free binary search for the event type
func (d *Dispatcher) findGroup(eventType uint32) any {
	reg := d.subs.Load()
	keys := reg.keys

	// Inlined binary search for better cache locality
	left, right := 0, len(keys)
	for left < right {
		mid := left + (right-left)/2
		if keys[mid] < eventType {
			left = mid + 1
		} else {
			right = mid
		}
	}

	if left < len(keys) && keys[left] == eventType {
		return reg.subs[left]
	}
	return nil
}

// Subscribe subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work.
func Subscribe[T Event](broker *Dispatcher, handler func(T)) context.CancelFunc {
	var event T
	return SubscribeTo(broker, event.Type(), handler)
}

// SubscribeTo subscribes to an event with the specified event type.
func SubscribeTo[T Event](broker *Dispatcher, eventType uint32, handler func(T)) context.CancelFunc {
	if broker.isClosed() {
		panic(errClosed)
	}

	broker.mu.Lock()
	defer broker.mu.Unlock()

	// Check if group already exists
	if existing := broker.findGroup(eventType); existing != nil {
		grp := groupOf[T](eventType, existing)
		sub := grp.Add(handler)
		return func() {
			grp.Del(sub)
		}
	}

	// Create new grp
	grp := &group[T]{cond: sync.NewCond(new(sync.Mutex))}
	sub := grp.Add(handler)

	// Copy-on-write with CAS loop: insert new entry in sorted position
	for {
		old := broker.subs.Load()
		idx := sort.Search(len(old.keys), func(i int) bool {
			return old.keys[i] >= eventType
		})

		// Create new arrays with space for one more element
		newKeys := make([]uint32, len(old.keys)+1)
		newSubs := make([]any, len(old.subs)+1)

		// Copy elements before insertion point
		copy(newKeys[:idx], old.keys[:idx])
		copy(newSubs[:idx], old.subs[:idx])

		// Insert new element
		newKeys[idx] = eventType
		newSubs[idx] = grp

		// Copy elements after insertion point
		copy(newKeys[idx+1:], old.keys[idx:])
		copy(newSubs[idx+1:], old.subs[idx:])

		// Atomically swap the registry
		newReg := &registry{keys: newKeys, subs: newSubs}
		if broker.subs.CompareAndSwap(old, newReg) {
			break
		}
	}

	// Start processing
	go grp.Process(broker.df, broker.done)
	return func() {
		grp.Del(sub)
	}
}

// Publish writes an event into the dispatcher
func Publish[T Event](broker *Dispatcher, ev T) {
	eventType := ev.Type()
	if sub := broker.findGroup(eventType); sub != nil {
		group := groupOf[T](eventType, sub)
		group.Broadcast(ev)
	}
}

// Count counts the number of subscribers, this is for testing only.
func (d *Dispatcher) count(eventType uint32) int {
	if group := d.findGroup(eventType); group != nil {
		return group.(interface{ Count() int }).Count()
	}
	return 0
}

// groupOf casts the subscriber group to the specified generic type
func groupOf[T Event](eventType uint32, subs any) *group[T] {
	if group, ok := subs.(*group[T]); ok {
		return group
	}

	panic(errConflict[T](eventType, subs))
}

// ------------------------------------- Subscriber -------------------------------------

// consumer represents a consumer with a message queue
type consumer[T Event] struct {
	queue []T  // Current work queue
	stop  bool // Stop signal
}

// Listen listens to the event queue and processes events
func (s *consumer[T]) Listen(c *sync.Cond, fn func(T)) {
	pending := make([]T, 0, 128)

	for {
		c.L.Lock()
		for len(s.queue) == 0 {
			switch {
			case s.stop:
				c.L.Unlock()
				return
			default:
				c.Wait()
			}
		}

		// Swap buffers and reset the current queue
		temp := s.queue
		s.queue = pending
		pending = temp
		s.queue = s.queue[:0]
		c.L.Unlock()

		// Outside of the critical section, process the work
		for i := 0; i < len(pending); i++ {
			fn(pending[i])
		}
	}
}

// ------------------------------------- Subscriber Group -------------------------------------

// group represents a consumer group
type group[T Event] struct {
	cond *sync.Cond
	subs []*consumer[T]
}

// Process periodically broadcasts events
func (s *group[T]) Process(interval time.Duration, done chan struct{}) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.cond.Broadcast()
		}
	}
}

// Broadcast sends an event to all consumers
func (s *group[T]) Broadcast(ev T) {
	s.cond.L.Lock()
	for _, sub := range s.subs {
		sub.queue = append(sub.queue, ev)
	}
	s.cond.L.Unlock()
}

// Add adds a subscriber to the list
func (s *group[T]) Add(handler func(T)) *consumer[T] {
	sub := &consumer[T]{
		queue: make([]T, 0, 128),
	}

	// Add the consumer to the list of active consumers
	s.cond.L.Lock()
	s.subs = append(s.subs, sub)
	s.cond.L.Unlock()

	// Start listening
	go sub.Listen(s.cond, handler)
	return sub
}

// Del removes a subscriber from the list
func (s *group[T]) Del(sub *consumer[T]) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	// Search and remove the subscriber
	sub.stop = true
	subs := make([]*consumer[T], 0, len(s.subs))
	for _, v := range s.subs {
		if v != sub {
			subs = append(subs, v)
		}
	}
	s.subs = subs
}

// ------------------------------------- Debugging -------------------------------------

var errClosed = fmt.Errorf("event dispatcher is closed")

// Count returns the number of subscribers in this group
func (s *group[T]) Count() int {
	return len(s.subs)
}

// String returns string representation of the type
func (s *group[T]) String() string {
	typ := reflect.TypeOf(s).String()
	idx := strings.LastIndex(typ, "/")
	typ = typ[idx+1 : len(typ)-1]
	return typ
}

// errConflict returns a conflict message
func errConflict[T any](eventType uint32, existing any) string {
	var want T
	return fmt.Sprintf(
		"conflicting event type, want=<%T>, registered=<%s>, event=0x%v",
		want, existing, eventType,
	)
}
