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

/*
cpu: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz
BenchmarkEvent/1-consumers-8         	10240418	       116.8 ns/op	  10240023 msg	       0 B/op	       0 allocs/op
BenchmarkEvent/10-consumers-8        	  923197	      1396 ns/op	   9231961 msg	       0 B/op	       0 allocs/op
BenchmarkEvent/100-consumers-8       	   97951	     12699 ns/op	   9795055 msg	       0 B/op	       0 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, subs := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("%d-consumers", subs), func(b *testing.B) {
			var count atomic.Int64
			d := NewDispatcher()
			for i := 0; i < subs; i++ {
				defer Subscribe(d, func(ev MyEvent1) {
					count.Add(1)
				})()
			}

			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				Publish(d, MyEvent1{})
			}

			messages := float64(count.Load())
			b.ReportMetric(messages, "ev")
			b.ReportMetric(messages/float64(b.N), "ev/op")
		})
	}
}

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
