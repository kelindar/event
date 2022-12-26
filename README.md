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

// Event type for testing purposes
type myEvent struct{
    Data string
}

// Type returns the event type
func (ev myEvent) Type() uint32 {
	return EventA
}
```

## Using Default Dispatcher

For convenience, this package provides a default global dispatcher that can be used with `On()` and `Emit()` package-level functions.

```go
// Subcribe to event A, and automatically unsubscribe at the end
defer event.On(func(e Event) {
    println("(consumer)", e.Data)
})()

// Publish few events
event.Emit(newEventA("event 1"))
event.Emit(newEventA("event 2"))
event.Emit(newEventA("event 3"))
```

## Using Specific Dispatcher

When publishing events, you can create a `Dispatcher` which is then used as a target of generic `event.Publish[T]()` and `event.Subscribe[T]()` functions to publish and subscribe to various event types respectively.

```go
bus := event.NewDispatcher()

// Subcribe to event A, and automatically unsubscribe at the end
defer event.Subscribe(bus, func(e Event) {
    println("(consumer 1)", e.Data)
})()

// Subcribe to event A, and automatically unsubscribe at the end
defer event.Subscribe(bus, func(e Event) {
    println("(consumer 2)", e.Data)
})()

// Publish few events
event.Publish(bus, newEventA("event 1"))
event.Publish(bus, newEventA("event 2"))
event.Publish(bus, newEventA("event 3"))
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
BenchmarkEvent/1x1-8        69634809       38.07 ns/op      26245794 ev/s        0 B/op     0 allocs/op
BenchmarkEvent/1x10-8       15903538       188.2 ns/op      53029988 ev/s       10 B/op     0 allocs/op
BenchmarkEvent/1x100-8       1345704        1679 ns/op      59502067 ev/s       68 B/op     0 allocs/op
BenchmarkEvent/10x1-8        6537975       504.9 ns/op      19796366 ev/s        0 B/op     0 allocs/op
BenchmarkEvent/10x10-8       1740412        1979 ns/op      50472108 ev/s       87 B/op     0 allocs/op
BenchmarkEvent/10x100-8       160111       17376 ns/op      57290578 ev/s     2616 B/op     0 allocs/op
```
