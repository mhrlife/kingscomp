package entity

import (
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"time"
)

type UserState struct {
	IsReady    bool `json:"isReady"`
	IsResigned bool `json:"isResigned"`

	LastAnsweredQuestionIndex int    `json:"lastAnsweredQuestionIndex"`
	DisplayName               string `json:"displayName"`
}
type GameInfo struct {
	CurrentQuestion          int              `json:"currentQuestion"`
	CurrentQuestionStartedAt time.Time        `json:"currentQuestionStartedAt"`
	CurrentQuestionEndsAt    time.Time        `json:"CurrentQuestionEndsAt"`
	CorrectAnswers           map[int64][]bool `json:"correctAnswers"`
}

type Lobby struct {
	ID            string  `json:"id"`
	Participants  []int64 `json:"participants"`
	CreatedAtUnix int64   `json:"created_at"`

	GameInfo  GameInfo   `json:"gameInfo"`
	Questions []Question `json:"questions"`

	UserState map[int64]UserState `json:"userState"`

	State string `json:"state"`
}

func (l Lobby) EveryoneReady() bool {
	return lo.Reduce(maps.Values(l.UserState), func(agg bool, item UserState, _ int) bool {
		if !agg {
			return agg
		}
		return item.IsResigned || item.IsReady
	}, true)
}

func (l Lobby) NotAnsweredUsers() []int64 {
	var userIds []int64
	for _, userId := range l.Participants {
		us := l.UserState[userId]
		if us.IsResigned {
			continue
		}
		if len(l.GameInfo.CorrectAnswers[userId]) != l.GameInfo.CurrentQuestion+1 {
			userIds = append(userIds, userId)
		}
	}
	return userIds
}

func (l Lobby) EntityID() ID {
	return NewID("lobby", l.ID)
}

type Question struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Answers       []string `json:"answers"`
	CorrectAnswer int      `json:"correctAnswer"`
}

func (q Question) EntityID() ID {
	return NewID("question", q.ID)
}
