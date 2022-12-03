// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package main

import (
	"time"

	"github.com/kelindar/event"
)

// Various event types
const EventA = 0x01

// Event type for testing purposes
type Event struct {
	Data string
}

// Type returns the event type
func (ev Event) Type() uint32 {
	return EventA
}

// newEventA creates a new instance of an event
func newEventA(data string) Event {
	return Event{Data: data}
}

func main() {
	bus := event.NewDispatcher()

	// Subcribe to event A, and automatically unsubscribe at the end
	defer event.SubscribeTo(bus, EventA, func(e Event) {
		println("(consumer 1)", e.Data)
	})()

	// Subcribe to event A, and automatically unsubscribe at the end
	defer event.SubscribeTo(bus, EventA, func(e Event) {
		println("(consumer 2)", e.Data)
	})()

	// Publish few events
	event.Publish(bus, newEventA("event 1"))
	event.Publish(bus, newEventA("event 2"))
	event.Publish(bus, newEventA("event 3"))

	time.Sleep(10 * time.Millisecond)
}
