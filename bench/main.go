package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/kelindar/bench"
	"github.com/kelindar/event"
)

func main() {
	bench.Run(func(b *bench.B) {
		for _, topics := range []int{1, 10} {
			for _, subs := range []int{1, 10, 100} {
				var count atomic.Int64
				var unsubscribers []func()

				// Create subscribers
				for i := 0; i < subs; i++ {
					for id := 10; id < 10+topics; id++ {
						unsub := event.OnType(uint32(id), func(ev Event) {
							count.Add(1)
						})
						unsubscribers = append(unsubscribers, unsub)
					}
				}

				bus := newChannelPubsub(topics, subs)

				b.RunN(fmt.Sprintf("%dx%d", topics, subs), func(i int) int {
					for id := 10; id < 10+topics; id++ {
						event.Emit(Event{ID: id})
					}
					return int(count.Swap(0))
				}, func(i int) int {
					for id := 10; id < 10+topics; id++ {
						bus.Emit(id, Event{ID: id})
					}
					return bus.Count()
				})

				// Cleanup subscribers after benchmark
				for _, unsub := range unsubscribers {
					unsub()
				}

				// Stop all channel subscribers
				bus.Close()
			}
		}

	}, bench.WithSamples(100), bench.WithReference())

}

type Event struct {
	ID int
}

func (t Event) Type() uint32 { return uint32(t.ID) }

type channelPubsub struct {
	wg    sync.WaitGroup
	count atomic.Int64
	subs  map[int][]chan Event
}

func newChannelPubsub(topics int, subs int) *channelPubsub {
	bus := &channelPubsub{
		subs: make(map[int][]chan Event, topics),
		wg:   sync.WaitGroup{},
	}

	for id := 10; id < 10+topics; id++ {
		subscribers := make([]chan Event, subs)
		for i := 0; i < subs; i++ {
			ch := make(chan Event, 1)
			subscribers[i] = ch

			// Start a subscriber
			bus.wg.Add(1)
			go func() {
				defer bus.wg.Done()
				for range ch {
					bus.count.Add(1)
				}
			}()
		}
		bus.subs[id] = subscribers
	}

	return bus
}

func (p *channelPubsub) Emit(id int, ev Event) {
	for _, sub := range p.subs[id] {
		sub <- ev
	}
}

func (p *channelPubsub) Count() int {
	return int(p.count.Swap(0))
}

func (p *channelPubsub) Close() {
	for _, subs := range p.subs {
		for _, sub := range subs {
			close(sub)
		}
	}

	p.wg.Wait()
}
