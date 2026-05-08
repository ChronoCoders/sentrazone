package control

import (
	"sync"

	"github.com/ChronoCoders/sentra/internal/models"
)

type EventBus struct {
	mu   sync.RWMutex
	subs []chan models.StatusEvent
	in   chan models.StatusEvent
}

func NewEventBus() *EventBus {
	b := &EventBus{
		in: make(chan models.StatusEvent, 100),
	}
	go b.dispatch()
	return b
}

func (b *EventBus) dispatch() {
	for event := range b.in {
		b.mu.RLock()
		for _, ch := range b.subs {
			select {
			case ch <- event:
			default:
			}
		}
		b.mu.RUnlock()
	}
}

func (b *EventBus) Publish(event models.StatusEvent) {
	select {
	case b.in <- event:
	default:
	}
}

func (b *EventBus) Subscribe() <-chan models.StatusEvent {
	ch := make(chan models.StatusEvent, 100)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch
}
