package gameserver

import (
	"github.com/google/uuid"
	"github.com/samber/lo"
	"kingscomp/internal/entity"
	"slices"
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
	info.Type = t
	e.mu.RLock()
	listeners := append(e.listeners[t], e.listeners[EventAny]...)
	for _, listener := range listeners {
		go listener.callback(info)
	}
	e.mu.RUnlock()
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
	EventAny       EventType = -1
	EventUserReady EventType = iota
	EventUserResigned
	EventJoinReminder
	EventLateResign
	EventForceLobbyReload
	EventUserAnswer
)

var eventTypes = map[EventType]string{
	EventUserReady:    "user-ready",
	EventUserResigned: "user-resigned",
	EventJoinReminder: "join-reminder",
	EventLateResign:   "late-resign",
}

func (e EventType) Type() string {
	t, ok := eventTypes[e]
	if !ok {
		return "undefined"
	}
	return t
}

type EventInfo struct {
	Type EventType

	AccountID int64
	Account   entity.Account

	QuestionIndex int
	UserAnswer    int
}

func (e EventInfo) IsType(acceptable ...EventType) bool {
	return isEvent(e.Type, acceptable...)
}

func isEvent(et EventType, acceptable ...EventType) bool {
	return slices.Contains(acceptable, et)
}
