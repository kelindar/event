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
go test -bench=. -benchmem -benchtime=10s
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/1x1-24                   814403188               14.72 ns/op             67.95 million/s        0 B/op          0 allocs/op
BenchmarkEvent/1x10-24                  161012098               84.61 ns/op             90.93 million/s      196 B/op          0 allocs/op
BenchmarkEvent/1x100-24                  7890922              1409 ns/op                70.95 million/s       10 B/op          0 allocs/op
BenchmarkEvent/10x1-24                  72358305               155.3 ns/op              64.38 million/s        0 B/op          0 allocs/op
BenchmarkEvent/10x10-24                  7632547              1315 ns/op                76.05 million/s       30 B/op          0 allocs/op
BenchmarkEvent/10x100-24                  832560             13541 ns/op                73.84 million/s      210 B/op          0 allocs/op
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

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkSubcribeConcurrent-24    	 1826686	       606.3 ns/op	    1648 B/op	       5 allocs/op
*/
func BenchmarkSubscribeConcurrent(b *testing.B) {
	d := NewDispatcher()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			unsub := Subscribe(d, func(ev MyEvent1) {})
			unsub()
		}
	})
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
