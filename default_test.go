// Copyright (c) Roman Atachiants and contributore. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for detaile.

package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
cpu: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz
BenchmarkEvent/1x1-8         	18729601	        59.74 ns/op	  16739743 ev/s	       0 B/op	       0 allocs/op
BenchmarkEvent/10x1-8        	  704184	      1749 ns/op	   5718222 ev/s	       0 B/op	       0 allocs/op
BenchmarkEvent/1x10-8        	 2838696	       357.5 ns/op	  27975523 ev/s	       0 B/op	       0 allocs/op
BenchmarkEvent/10x10-8       	  179293	     12922 ns/op	   7738928 ev/s	       1 B/op	       0 allocs/op
BenchmarkEvent/1x100-8       	  249906	      4945 ns/op	  20223782 ev/s	       9 B/op	       0 allocs/op
BenchmarkEvent/10x100-8      	   51558	     38953 ns/op	  25657300 ev/s	     257 B/op	       0 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, subs := range []int{1, 10, 100} {
		for _, topics := range []int{1, 10} {
			b.Run(fmt.Sprintf("%dx%d", topics, subs), func(b *testing.B) {
				var count atomic.Int64
				for i := 0; i < subs; i++ {
					for id := 0; id < topics; id++ {
						defer OnType(uint32(id), func(ev MyEvent3) {
							count.Add(1)
						})()
					}
				}

				start := time.Now()
				b.ReportAllocs()
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					for id := 0; id < topics; id++ {
						Emit(MyEvent3{ID: id})
					}
				}

				elapsed := time.Since(start)
				rate := float64(count.Load()) / elapsed.Seconds()
				b.ReportMetric(rate, "ev/s")
			})
		}
	}
}

func TestDefaultPublish(t *testing.T) {
	var wg sync.WaitGroup

	// Subscribe
	var count int64
	defer On(func(ev MyEvent1) {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})()

	defer OnType(TypeEvent1, func(ev MyEvent1) {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})()

	// Publish
	wg.Add(4)
	Emit(MyEvent1{})
	Emit(MyEvent1{})

	// Wait and check
	wg.Wait()
	assert.Equal(t, int64(4), count)
}
