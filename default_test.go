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
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/1x1-24         	38709926	        31.94 ns/op	        30.89 million/s	       1 B/op	       0 allocs/op
BenchmarkEvent/1x10-24        	 8107938	       133.7 ns/op	        74.76 million/s	      45 B/op	       0 allocs/op
BenchmarkEvent/1x100-24       	  774168	      1341 ns/op	        72.65 million/s	     373 B/op	       0 allocs/op
BenchmarkEvent/10x1-24        	 5755402	       301.1 ns/op	        32.98 million/s	       7 B/op	       0 allocs/op
BenchmarkEvent/10x10-24       	  750022	      1503 ns/op	        64.47 million/s	     438 B/op	       0 allocs/op
BenchmarkEvent/10x100-24      	   69363	     14878 ns/op	        67.11 million/s	    3543 B/op	       0 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, topics := range []int{1, 10} {
		for _, subs := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("%dx%d", topics, subs), func(b *testing.B) {
				var count atomic.Int64
				for i := 0; i < subs; i++ {
					for id := 10; id < 10+topics; id++ {
						defer OnType(uint32(id), func(ev MyEvent3) {
							count.Add(1)
						})()
					}
				}

				start := time.Now()
				b.ReportAllocs()
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					for id := 10; id < 10+topics; id++ {
						Emit(MyEvent3{ID: id})
					}
				}

				elapsed := time.Since(start)
				rate := float64(count.Load()) / 1e6 / elapsed.Seconds()
				b.ReportMetric(rate, "million/s")
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
