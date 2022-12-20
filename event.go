// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Event represents an event contract
type Event interface {
	Type() uint32
}

// ------------------------------------- Dispatcher -------------------------------------

// Dispatcher represents an event dispatcher.
type Dispatcher struct {
	subs sync.Map
}

// NewDispatcher creates a new dispatcher of events.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Subscribe subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work.
func Subscribe[T Event](broker *Dispatcher, handler func(T)) context.CancelFunc {
	var event T
	return SubscribeTo(broker, event.Type(), handler)
}

// SubscribeTo subscribes to an event with the specified event type.
func SubscribeTo[T Event](broker *Dispatcher, eventType uint32, handler func(T)) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	sub := &consumer[T]{
		exec: handler,
	}

	// Add to consumer group, if it doesn't exist it will create one
	s, _ := broker.subs.LoadOrStore(eventType, &group[T]{
		cond: sync.NewCond(new(sync.Mutex)),
	})
	group := groupOf[T](eventType, s)
	group.Add(ctx, sub)

	// Return unsubscribe function
	return func() {
		group.Del(sub)
		cancel() // Stop async processing
	}
}

// Publish writes an event into the dispatcher
func Publish[T Event](broker *Dispatcher, ev T) {
	if s, ok := broker.subs.Load(ev.Type()); ok {
		group := groupOf[T](ev.Type(), s)
		group.Broadcast(ev)
	}
}

// Count counts the number of subscribers, this is for testing only.
func (d *Dispatcher) count(eventType uint32) int {
	if group, ok := d.subs.Load(eventType); ok {
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
	cond  *sync.Cond
	queue []T
	exec  func(T) // Process callback
}

// Listen listens to the event queue and processes events
func (s *consumer[T]) Listen(c *sync.Cond) {
	s.cond = c
	work := make([]T, 0, 128)

	for {
		c.L.Lock()
		for len(s.queue) == 0 {
			s.cond.Wait()
		}

		// Grow the work buffer if needed
		if cap(work) < len(s.queue) {
			work = make([]T, len(s.queue))
		}

		// Copy the work we need to do into the work buffer
		work = work[:len(s.queue)]
		copy(work, s.queue)
		s.queue = s.queue[:0]
		c.L.Unlock()

		// Outside of the critical section, process the work
		for i := 0; i < len(work); i++ {
			s.exec(work[i])
		}
	}
}

// ------------------------------------- Subscriber Group -------------------------------------

// group represents a consumer group
type group[T Event] struct {
	cond *sync.Cond
	subs []*consumer[T]
}

// Broadcast sends an event to all consumers
func (s *group[T]) Broadcast(ev T) {
	s.cond.L.Lock()
	for _, sub := range s.subs {
		sub.queue = append(sub.queue, ev)
	}
	s.cond.L.Unlock()
	s.cond.Broadcast()
}

// Add adds a subscriber to the list
func (s *group[T]) Add(ctx context.Context, sub *consumer[T]) {
	go sub.Listen(s.cond)

	// Add the consumer to the list of active consumers
	s.cond.L.Lock()
	s.subs = append(s.subs, sub)
	s.cond.L.Unlock()
}

// Del removes a subscriber from the list
func (s *group[T]) Del(sub *consumer[T]) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	// Search and remove the subscriber
	subs := make([]*consumer[T], 0, len(s.subs))
	for _, v := range s.subs {
		if v != sub {
			subs = append(subs, v)
		}
	}
	s.subs = subs
}

// ------------------------------------- Debugging -------------------------------------

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
