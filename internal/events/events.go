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

type Queue interface {
	Dispatch(ctx context.Context, t EventType, info EventInfo) error
	Register(t EventType, callback Callback) (func(), error)
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
	EventNewScore
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
	EventNewScore:         "new-score",
}

func (e EventType) Type() string {
	t, ok := eventTypes[e]
	if !ok {
		return "undefined"
	}
	return t
}

type EventInfo struct {
	UUID string    `json:"uuid,omitempty"`
	Type EventType `json:"type,omitempty"`

	// user information
	AccountID int64          `json:"account_id,omitempty"`
	Account   entity.Account `json:"account,omitempty"`

	// lobby information
	LobbyID string `json:"lobby_id"`

	// answer information
	QuestionIndex int `json:"question_index,omitempty"`
	UserAnswer    int `json:"user_answer,omitempty"`

	// telegram information
	Message *telebot.Message `json:"message,omitempty"`

	// leaderboard information
	Score int64 `json:"score,omitempty"`
}

func (e EventInfo) IsType(acceptable ...EventType) bool {
	return isEvent(e.Type, acceptable...)
}

func isEvent(et EventType, acceptable ...EventType) bool {
	return slices.Contains(acceptable, et)
}
