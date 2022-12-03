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
		queue: make(chan T, 1024),
		exec:  handler,
	}

	// Add to consumer group, if it doesn't exist it will create one
	s, _ := broker.subs.LoadOrStore(eventType, new(group[T]))
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

// ------------------------------------- Subscriber List -------------------------------------

// consumer represents a consumer with a message queue
type consumer[T Event] struct {
	queue chan T  // Message buffer
	exec  func(T) // Process callback
}

// Listen listens to the event queue and processes events
func (s *consumer[T]) Listen(ctx context.Context) {
	for {
		select {
		case ev := <-s.queue:
			s.exec(ev)
		case <-ctx.Done():
			return
		}
	}
}

// group represents a consumer group
type group[T Event] struct {
	sync.RWMutex
	subs []*consumer[T]
}

// Broadcast sends an event to all consumers
func (s *group[T]) Broadcast(ev T) {
	s.RLock()
	defer s.RUnlock()
	for _, sub := range s.subs {
		sub.queue <- ev
	}
}

// Add adds a subscriber to the list
func (s *group[T]) Add(ctx context.Context, sub *consumer[T]) {
	go sub.Listen(ctx)

	// Add the consumer to the list of active consumers
	s.Lock()
	s.subs = append(s.subs, sub)
	s.Unlock()
}

// Del removes a subscriber from the list
func (s *group[T]) Del(sub *consumer[T]) {
	s.Lock()
	defer s.Unlock()

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
