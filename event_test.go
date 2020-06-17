package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	ev := new(Bus)
	assert.NotNil(t, ev)

	var count int
	cancel := ev.On("event1", func(v interface{}) {
		count += v.(int)
	})
	defer cancel()

	ev.Notify("event1", 1)
	ev.Notify("event2", 2)
	ev.Notify("event1", 1)
	ev.Notify("event1", 1)
	ev.Notify("event3", 3)

	for count < 3 {
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, 3, count)
	cancel()

	ev.Notify("event2", 2)
	assert.Equal(t, 3, count)
}
