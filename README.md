# Event Bus

This repository contains a simple in-process event bus to be used to decouple internal modules. 

```
var ev event.Bus

// Subscribe to events 
var count int
cancel := ev.On("event1", func(v interface{}) {
    count += v.(int)
})

// Notify events
ev.Notify("event1", 1)
ev.Notify("event1", 1)

// Unsubscribe from "event1"
cancel()
```