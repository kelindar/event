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
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/1x1-24          164887946       14.55 ns/op     68.70 million/s      0 B/op     0 allocs/op
BenchmarkEvent/1x10-24          28896586       80.06 ns/op     88.02 million/s    204 B/op     0 allocs/op
BenchmarkEvent/1x100-24          1535168        1397 ns/op     71.57 million/s     26 B/op     0 allocs/op
BenchmarkEvent/10x1-24          14288467       256.1 ns/op     39.03 million/s      0 B/op     0 allocs/op
BenchmarkEvent/10x10-24          1624722        1265 ns/op     78.59 million/s     65 B/op     0 allocs/op
BenchmarkEvent/10x100-24          164623       12767 ns/op     78.33 million/s    456 B/op     0 allocs/op
```
