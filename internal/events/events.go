package events

import (
	"context"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
	"slices"
)

type PubSub interface {
	Dispatch(ctx context.Context, key string, t EventType, info EventInfo) error
	Register(key string, t EventType, callback Callback) (func(), error)
	Clean(key string, t EventType) error
	Close() error
}

type Eventer interface {
	Dispatch(t EventType, info EventInfo) error
	Clean(t EventType) error
	Close() error
	Register(t EventType, callback Callback) (func() int, error)
}

type EventType int

type Callback func(info EventInfo)

type Listener struct {
	callback Callback
	uuid     string
}

const (
	EventAny       EventType = -1
	EventUserReady EventType = iota
	EventUserResigned
	EventJoinReminder
	EventLateResign
	EventForceLobbyReload
	EventUserAnswer
	EventGameClosed
)

var eventTypes = map[EventType]string{
	EventUserReady:        "user-ready",
	EventUserResigned:     "user-resigned",
	EventJoinReminder:     "join-reminder",
	EventLateResign:       "late-resign",
	EventForceLobbyReload: "lobby-reload",
	EventUserAnswer:       "user-answer",
	EventGameClosed:       "event-game-closed",
	EventAny:              "any",
}

func (e EventType) Type() string {
	t, ok := eventTypes[e]
	if !ok {
		return "undefined"
	}
	return t
}

type EventInfo struct {
	Type EventType `json:"type"`

	AccountID int64          `json:"accountID"`
	Account   entity.Account `json:"account"`

	QuestionIndex int `json:"questionIndex"`
	UserAnswer    int `json:"userAnswer"`

	UUID    string           `json:"uuid"`
	Message *telebot.Message `json:"message"`
}

func (e EventInfo) IsType(acceptable ...EventType) bool {
	return isEvent(e.Type, acceptable...)
}

func isEvent(et EventType, acceptable ...EventType) bool {
	return slices.Contains(acceptable, et)
}
