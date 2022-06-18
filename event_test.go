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
BenchmarkEvent/1-consumers-8         	10021444	       119.1 ns/op	  10021301 msg	       0 B/op	       0 allocs/op
BenchmarkEvent/10-consumers-8        	  799999	      1595 ns/op	   7999915 msg	       0 B/op	       0 allocs/op
BenchmarkEvent/100-consumers-8       	   99048	     14308 ns/op	   9904769 msg	       0 B/op	       0 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, subs := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("%d-consumers", subs), func(b *testing.B) {
			var count uint64
			d := NewDispatcher[testEvent]()
			for i := 0; i < subs; i++ {
				defer d.Subscribe(TestEventType, func(ev testEvent) {
					atomic.AddUint64(&count, 1)
				})()
			}

			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				d.Publish(testEvent{})
			}
			b.ReportMetric(float64(count), "msg")
		})
	}
}

func TestPublish(t *testing.T) {
	d := NewDispatcher[testEvent]()
	var wg sync.WaitGroup

	// Subscribe
	var count int64
	defer d.Subscribe(TestEventType, func(ev testEvent) {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})()

	// Publish
	wg.Add(2)
	d.Publish(testEvent{})
	d.Publish(testEvent{})

	// Wait and check
	wg.Wait()
	assert.Equal(t, int64(2), count)
}

func TestUnsubscribe(t *testing.T) {
	d := NewDispatcher[testEvent]()
	unsubscribe := d.Subscribe(TestEventType, func(ev testEvent) {
		// Nothing
	})

	assert.Equal(t, 1, d.count(TestEventType))
	unsubscribe()
	assert.Equal(t, 0, d.count(TestEventType))
}

func TestConcurrent(t *testing.T) {
	const max = 1000000
	var count int64
	var wg sync.WaitGroup
	wg.Add(1)

	d := NewDispatcher[testEvent]()
	defer d.Subscribe(TestEventType, func(ev testEvent) {
		if current := atomic.AddInt64(&count, 1); current == max {
			wg.Done()
		}
	})()

	// Asynchronously publish
	go func() {
		for i := 0; i < max; i++ {
			d.Publish(testEvent{})
		}
	}()

	defer d.Subscribe(TestEventType, func(ev testEvent) {
		// Subscriber that does nothing
	})()

	wg.Wait()
	assert.Equal(t, max, int(count))
}

// ------------------------------------- Test Event -------------------------------------

const TestEventType = 0xff

type testEvent struct{}

func (testEvent) Type() uint32 {
	return TestEventType
}
