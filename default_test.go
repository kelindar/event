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
BenchmarkEvent/1x1-8         	19471857	        61.37 ns/op	  16295560 ev/s	       0 B/op	       0 allocs/op
BenchmarkEvent/10x1-8        	  701856	      1772 ns/op	   5643499 ev/s	       0 B/op	       0 allocs/op
BenchmarkEvent/1x10-8        	 4033690	       389.9 ns/op	  25646847 ev/s	       0 B/op	       0 allocs/op
BenchmarkEvent/10x10-8       	  107130	     12355 ns/op	   8093423 ev/s	       2 B/op	       0 allocs/op
BenchmarkEvent/1x100-8       	  378753	      3680 ns/op	  27177107 ev/s	      16 B/op	       0 allocs/op
BenchmarkEvent/10x100-8      	   28020	     46035 ns/op	  21738149 ev/s	     308 B/op	       0 allocs/op
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
