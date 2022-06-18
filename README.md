<p align="center">
<img width="330" height="110" src=".github/logo.png" border="0" alt="kelindar/event">
<br>
<img src="https://img.shields.io/github/go-mod/go-version/kelindar/event" alt="Go Version">
<a href="https://pkg.go.dev/github.com/kelindar/event"><img src="https://pkg.go.dev/badge/github.com/kelindar/event" alt="PkgGoDev"></a>
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
<a href="https://coveralls.io/github/kelindar/event"><img src="https://coveralls.io/repos/github/kelindar/event/badge.svg" alt="Coverage"></a>
</p>

## Generic In-Process Pub/Sub

This repository contains a **simple, in-process event dispatcher** to be used to decouple internal modules. It provides a generic way to define events, publish and subscribe to them.

```go
// Various event types
const EventA = 0x01
const EventB = 0x02

// Event type for testing purposes
type myEvent struct{
    kind uint32
    Data string
}

// Type returns the event type
func (ev myEvent) Type() uint32 {
	return ev.kind
}
```

When publishing events, you can create a `Dispatcher[T]` which allows to `Publish()` and `Subscribe()` to various event types.

```go
bus := event.NewDispatcher[Event]()

// Subcribe to event A, and automatically unsubscribe at the end
defer bus.Subscribe(EventA, func(e Event) {
    println("(consumer 1)", e.Data)
})()

// Subcribe to event A, and automatically unsubscribe at the end
defer bus.Subscribe(EventA, func(e Event) {
    println("(consumer 2)", e.Data)
})()

// Publish few events
bus.Publish(newEventA("event 1"))
bus.Publish(newEventA("event 2"))
bus.Publish(newEventA("event 3"))
```

It should output something along these lines, where order is not guaranteed given that both subscribers are processing messages asyncrhonously.

```
(consumer 2) event 1
(consumer 2) event 2
(consumer 2) event 3
(consumer 1) event 1
(consumer 1) event 2
(consumer 1) event 3
```

## Benchmarks

```
cpu: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz
BenchmarkEvent/1-consumers-8   	10021444   119.1 ns/op  10021301 msg   0 B/op   0 allocs/op
BenchmarkEvent/10-consumers-8  	  799999    1595 ns/op   7999915 msg   0 B/op   0 allocs/op
BenchmarkEvent/100-consumers-8 	   99048   14308 ns/op   9904769 msg   0 B/op   0 allocs/op
```
