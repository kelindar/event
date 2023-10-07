// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	d := NewDispatcher()
	var wg sync.WaitGroup

	// Subscribe, must be received in order
	var count int64
	defer Subscribe(d, func(ev MyEvent1) {
		assert.Equal(t, int(atomic.AddInt64(&count, 1)), ev.Number)
		wg.Done()
	})()

	// Publish
	wg.Add(3)
	Publish(d, MyEvent1{Number: 1})
	Publish(d, MyEvent1{Number: 2})
	Publish(d, MyEvent1{Number: 3})

	// Wait and check
	wg.Wait()
	assert.Equal(t, int64(3), count)
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

func TestCloseDispatcher(t *testing.T) {
	d := NewDispatcher()
	defer SubscribeTo(d, TypeEvent1, func(ev MyEvent2) {})()

	assert.NoError(t, d.Close())
	assert.Panics(t, func() {
		SubscribeTo(d, TypeEvent1, func(ev MyEvent2) {})
	})
}

func TestMatrix(t *testing.T) {
	const amount = 1000
	for _, subs := range []int{1, 10, 100} {
		for _, topics := range []int{1, 10} {
			expected := subs * topics * amount
			t.Run(fmt.Sprintf("%dx%d", topics, subs), func(t *testing.T) {
				var count atomic.Int64
				var wg sync.WaitGroup
				wg.Add(expected)

				d := NewDispatcher()
				for i := 0; i < subs; i++ {
					for id := 0; id < topics; id++ {
						defer SubscribeTo(d, uint32(id), func(ev MyEvent3) {
							count.Add(1)
							wg.Done()
						})()
					}
				}

				for n := 0; n < amount; n++ {
					for id := 0; id < topics; id++ {
						go Publish(d, MyEvent3{ID: id})
					}
				}

				wg.Wait()
				assert.Equal(t, expected, int(count.Load()))
			})
		}
	}
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

type MyEvent3 struct {
	ID int
}

func (t MyEvent3) Type() uint32 { return uint32(t.ID) }
