// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import "context"

// defaultDispatcher initializes a default in-process dispatcher
var defaultDispatcher = NewDispatcher()

// On subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work. This
// functions same way as Subscribe() but uses the default dispatcher instead.
func On[T Event](handler func(T)) context.CancelFunc {
	return Subscribe(defaultDispatcher, handler)
}

// OnType subscribes to an event with the specified event type. This functions
// same way as SubscribeTo() but uses the default dispatcher instead.
func OnType[T Event](eventType uint32, handler func(T)) context.CancelFunc {
	return SubscribeTo(defaultDispatcher, eventType, handler)
}

// Emit writes an event into the dispatcher. This functions same way as
// Publish() but uses the default dispatcher instead.
func Emit[T Event](ev T) {
	Publish(defaultDispatcher, ev)
}
