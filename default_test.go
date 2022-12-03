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
BenchmarkDefault/1-consumers-8         	 9965634	       120.6 ns/op	   9965601 msg	       0 B/op	       0 allocs/op
BenchmarkDefault/10-consumers-8        	  800031	      1365 ns/op	   8000305 msg	       0 B/op	       0 allocs/op
BenchmarkDefault/100-consumers-8       	   82742	     15891 ns/op	   8274183 msg	       0 B/op	       0 allocs/op
*/
func BenchmarkDefault(b *testing.B) {
	for _, subs := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("%d-consumers", subs), func(b *testing.B) {
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
			b.ReportMetric(float64(count), "msg")
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
