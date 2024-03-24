package entity

import (
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"sort"
	"time"
)

type UserState struct {
	IsReady    bool `json:"isReady"`
	IsResigned bool `json:"isResigned"`

	LastAnsweredQuestionIndex int    `json:"lastAnsweredQuestionIndex"`
	DisplayName               string `json:"displayName"`
}
type GameInfo struct {
	CurrentQuestion          int                `json:"currentQuestion"`
	CurrentQuestionStartedAt time.Time          `json:"currentQuestionStartedAt"`
	CurrentQuestionEndsAt    time.Time          `json:"CurrentQuestionEndsAt"`
	CorrectAnswers           map[int64][]Answer `json:"correctAnswers"`
}

type Answer struct {
	Correct  bool          `json:"correct"`
	Duration time.Duration `json:"duration"`
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

type Score struct {
	AccountID   int64  `json:"accountID"`
	Score       int64  `json:"score"`
	DisplayName string `json:"displayName"`
}

func (l Lobby) Scores() []Score {
	var scores []Score
	for _, accountId := range l.Participants {
		score := lo.Reduce(l.GameInfo.CorrectAnswers[accountId], func(agg int64, item Answer, i int) int64 {
			if !item.Correct {
				return agg
			}
			agg += 10 // for correct answer
			questionDuration := l.GameInfo.CurrentQuestionEndsAt.Sub(l.GameInfo.CurrentQuestionStartedAt).Seconds()
			agg += 5 - int64(item.Duration.Seconds()/questionDuration*5) // for earlier answers
			return agg
		}, 0)
		scores = append(scores, Score{
			DisplayName: l.UserState[accountId].DisplayName,
			Score:       score,
			AccountID:   accountId,
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores
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
