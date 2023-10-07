// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	d := NewDispatcher()
	var wg sync.WaitGroup

	// Subscribe
	var count int64
	defer Subscribe(d, func(ev MyEvent1) {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})()

	// Publish
	wg.Add(2)
	Publish(d, MyEvent1{})
	Publish(d, MyEvent1{})

	// Wait and check
	wg.Wait()
	assert.Equal(t, int64(2), count)
}

func TestUnsubscribe(t *testing.T) {
	d := NewDispatcher()
	assert.Equal(t, 0, d.count(TypeEvent1))
	unsubscribe := Subscribe(d, func(ev MyEvent1) {
		// Nothing
	})

	assert.Equal(t, 1, d.count(TypeEvent1))
	unsubscribe()
	assert.Equal(t, 0, d.count(TypeEvent1))
}

func TestConcurrent(t *testing.T) {
	const max = 1000000
	var count int64
	var wg sync.WaitGroup
	wg.Add(1)

	d := NewDispatcher()
	defer Subscribe(d, func(ev MyEvent1) {
		if current := atomic.AddInt64(&count, 1); current == max {
			wg.Done()
		}
	})()

	// Asynchronously publish
	go func() {
		for i := 0; i < max; i++ {
			Publish(d, MyEvent1{})
		}
	}()

	defer Subscribe(d, func(ev MyEvent1) {
		// Subscriber that does nothing
	})()

	wg.Wait()
	assert.Equal(t, max, int(count))
}

func TestSubscribeDifferentType(t *testing.T) {
	d := NewDispatcher()
	assert.Panics(t, func() {
		SubscribeTo(d, TypeEvent1, func(ev MyEvent1) {})
		SubscribeTo(d, TypeEvent1, func(ev MyEvent2) {})
	})
}

func TestPublishDifferentType(t *testing.T) {
	d := NewDispatcher()
	assert.Panics(t, func() {
		SubscribeTo(d, TypeEvent1, func(ev MyEvent2) {})
		Publish(d, MyEvent1{})
	})
}

// ------------------------------------- Test Events -------------------------------------

const (
	TypeEvent1 = 0x1
	TypeEvent2 = 0x2
)

type MyEvent1 struct {
	Number int
}

func (t MyEvent1) Type() uint32 { return TypeEvent1 }

type MyEvent2 struct {
	Text string
}

func (t MyEvent2) Type() uint32 { return TypeEvent2 }
