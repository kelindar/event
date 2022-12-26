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
BenchmarkEvent/1x1-8         	22845045	        48.90 ns/op	  20399340 ev/s	       1 B/op	       0 allocs/op
BenchmarkEvent/1x10-8        	 5448380	       212.2 ns/op	  46183805 ev/s	      47 B/op	       0 allocs/op
BenchmarkEvent/1x100-8       	  631585	      1876 ns/op	  50592945 ev/s	     377 B/op	       0 allocs/op
BenchmarkEvent/10x1-8        	 2284117	       503.2 ns/op	  19805320 ev/s	      11 B/op	       0 allocs/op
BenchmarkEvent/10x10-8       	  521727	      2041 ns/op	  47717331 ev/s	     373 B/op	       0 allocs/op
BenchmarkEvent/10x100-8      	   68062	     17626 ns/op	  55497296 ev/s	    2707 B/op	       0 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, topics := range []int{1, 10} {
		for _, subs := range []int{1, 10, 100} {
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
