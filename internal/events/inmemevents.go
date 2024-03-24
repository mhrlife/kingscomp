package events

import (
	"context"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"sync"
)

// InMemoryEvents todo: needs to be scalable, use redis instead
type InMemoryEvents struct {
	mu        sync.RWMutex
	listeners map[EventType][]Listener
}

func NewInMemoryEvents() *InMemoryEvents {
	return &InMemoryEvents{
		listeners: make(map[EventType][]Listener),
	}
}

func (e *InMemoryEvents) Dispatch(t EventType, info EventInfo) error {
	info.Type = t
	e.mu.RLock()
	var add []Listener
	if !info.IsType(EventAny) {
		add = append(add, e.listeners[EventAny]...)
	}
	listeners := append(e.listeners[t], add...)
	logrus.WithFields(logrus.Fields{
		"count": len(listeners),
		"event": t.Type(),
	}).Info("started dispatching event")
	for _, listener := range listeners {
		go listener.callback(info)
	}
	logrus.WithFields(logrus.Fields{
		"count": len(listeners),
		"event": t.Type(),
	}).Info("dispatch done")
	e.mu.RUnlock()
	return nil
}

func (e *InMemoryEvents) listenerCount(t EventType) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.listeners[t])
}

func (e *InMemoryEvents) Clean(t EventType) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners[t] = make([]Listener, 0)
	return nil
}

func (e *InMemoryEvents) Close() error {
	if e.mu.TryLock() {
		defer e.mu.Unlock()
	}

	e.listeners = make(map[EventType][]Listener)
	return nil
}

func (e *InMemoryEvents) Register(t EventType, callback Callback) (func() int, error) {
	e.mu.Lock()

	uid := uuid.New().String()
	e.listeners[t] = append(e.listeners[t], Listener{
		callback: callback,
		uuid:     uid,
	})

	e.mu.Unlock()

	return func() int {
		e.mu.Lock()
		defer e.mu.Unlock()

		e.listeners[t] = lo.Filter(e.listeners[t], func(item Listener, index int) bool {
			return uid != item.uuid
		})
		l := len(e.listeners[t])
		return l

	}, nil
}

var _ PubSub = &InMemPubSub{}

type InMemPubSub struct {
	keys sync.Map
}

func NewInMemPubSub() *InMemPubSub {
	return &InMemPubSub{}
}

func (i *InMemPubSub) Dispatch(ctx context.Context, key string, t EventType, info EventInfo) error {
	iInMem, _ := i.keys.LoadOrStore(key, NewInMemoryEvents())
	return iInMem.(*InMemoryEvents).Dispatch(t, info)
}

func (i *InMemPubSub) Register(key string, t EventType, callback Callback) (func(), error) {
	iInMem, _ := i.keys.LoadOrStore(key, NewInMemoryEvents())
	cancel, _ := iInMem.(*InMemoryEvents).Register(t, callback)
	return func() {
		if cancel() == 0 {
			i.keys.Delete(key)
		}
	}, nil
}

func (i *InMemPubSub) Clean(key string, t EventType) error {
	iInMem, _ := i.keys.LoadOrStore(key, NewInMemoryEvents())
	return iInMem.(*InMemoryEvents).Clean(t)
}

func (i *InMemPubSub) Close() error {
	return nil
}
