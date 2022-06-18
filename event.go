// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import (
	"context"
	"sync"
)

// Event represents an event contract
type Event interface {
	Type() uint32
}

// ------------------------------------- Dispatcher -------------------------------------

// Dispatcher represents an event dispatcher.
type Dispatcher[T Event] struct {
	subs sync.Map
}

// NewDispatcher creates a new dispatcher of events.
func NewDispatcher[T Event]() *Dispatcher[T] {
	return &Dispatcher[T]{}
}

// loadOrStore finds a subscriber group or creates a new one
func (d *Dispatcher[T]) loadOrStore(key uint32) *group[T] {
	s, _ := d.subs.LoadOrStore(key, new(group[T]))
	return s.(*group[T])
}

// Publish writes an event into the dispatcher
func (d *Dispatcher[T]) Publish(ev T) {
	if g, ok := d.subs.Load(ev.Type()); ok {
		g.(*group[T]).Broadcast(ev)
	}
}

// Subscribe subscribes to an callback event
func (d *Dispatcher[T]) Subscribe(eventType uint32, handler func(T)) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	sub := &consumer[T]{
		queue: make(chan T, 1024),
		exec:  handler,
	}

	// Add to consumer group, if it doesn't exist it will create one
	group := d.loadOrStore(eventType)
	group.Add(ctx, sub)

	// Return unsubscribe function
	return func() {
		d.unsubscribe(eventType, sub) // Remove from the list
		cancel()                      // Stop async processing
	}
}

// Count counts the number of subscribers
func (d *Dispatcher[T]) count(eventType uint32) int {
	return len(d.loadOrStore(eventType).subs)
}

// unsubscribe removes the subscriber from the list of subscribers
func (d *Dispatcher[T]) unsubscribe(eventType uint32, sub *consumer[T]) {
	group := d.loadOrStore(eventType)
	group.Del(sub)
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
