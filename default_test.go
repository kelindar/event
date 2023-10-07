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
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEmit/1-subs-24         	13407880	        87.10 ns/op	        13.41 million	       0 B/op	       0 allocs/op
BenchmarkEmit/10-subs-24        	 1000000	      1012 ns/op	        10.00 million	       0 B/op	       0 allocs/op
BenchmarkEmit/100-subs-24       	  103896	     11714 ns/op	        10.39 million	       0 B/op	       0 allocs/op
*/
func BenchmarkEmit(b *testing.B) {
	for _, subs := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("%d-subs", subs), func(b *testing.B) {
			var count uint64
			for i := 0; i < subs; i++ {
				defer On(func(ev MyEvent1) {
					atomic.AddUint64(&count, 1)
				})()
			}

			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				Emit(MyEvent1{})
			}
			b.ReportMetric(float64(count)/1e6, "million")
		})
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
