package gameserver

import (
	"github.com/google/uuid"
	"github.com/samber/lo"
	"kingscomp/internal/entity"
	"sync"
)

type EventType int

type Callback func(info EventInfo)

type Listener struct {
	callback Callback
	uuid     string
}

type Events struct {
	mu        sync.RWMutex
	listeners map[EventType][]Listener
}

func NewEvents() *Events {
	return &Events{
		listeners: make(map[EventType][]Listener),
	}
}

func (e *Events) Dispatch(t EventType, info EventInfo) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, listener := range e.listeners[t] {
		go listener.callback(info)
	}
}

func (e *Events) ListenerCount(t EventType) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.listeners[t])
}

func (e *Events) Clean(t EventType) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners[t] = make([]Listener, 0)
}

func (e *Events) Register(t EventType, callback Callback) func() {
	e.mu.Lock()

	uid := uuid.New().String()
	e.listeners[t] = append(e.listeners[t], Listener{
		callback: callback,
		uuid:     uid,
	})

	e.mu.Unlock()

	return func() {
		e.mu.Lock()
		e.listeners[t] = lo.Filter(e.listeners[t], func(item Listener, index int) bool {
			return uid != item.uuid
		})
		e.mu.Unlock()
	}
}

const (
	EventReady EventType = iota
	EventUserResigned
	EventJoinReminder
	EventLateResign
)

type EventInfo struct {
	AccountID int64
	Account   entity.Account
}
