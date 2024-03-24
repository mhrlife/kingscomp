package webapp

import (
	"github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/entity"
	"kingscomp/internal/events"
	"strconv"
)

type AnswerSerializer struct {
	Correct bool `json:"correct"`
	Took    int  `json:"took"` // in seconds
}

func NewAnswerSerializer(answer entity.Answer) AnswerSerializer {
	return AnswerSerializer{
		Correct: answer.Correct,
		Took:    int(answer.Duration.Seconds()),
	}
}

type ParticipantAnswerHistory struct {
	AnswerHistory []AnswerSerializer `json:"answerHistory"`
}
type ParticipantSerializer struct {
	ID          int64                    `json:"id"`
	DisplayName string                   `json:"displayName"`
	IsReady     bool                     `json:"isReady"`
	IsResigned  bool                     `json:"isResigned"`
	History     ParticipantAnswerHistory `json:"history"`
}

func NewParticipantSerializer(state entity.UserState, id int64, answers []entity.Answer) ParticipantSerializer {
	if answers == nil {
		answers = make([]entity.Answer, 0)
	}
	return ParticipantSerializer{
		ID:          id,
		DisplayName: state.DisplayName,
		IsReady:     state.IsReady,
		IsResigned:  state.IsResigned,
		History: ParticipantAnswerHistory{
			AnswerHistory: lo.Map(answers, func(item entity.Answer, _ int) AnswerSerializer {
				return NewAnswerSerializer(item)
			}),
		},
	}
}

type QuestionSerializer struct {
	Index    int      `json:"index"`
	Question string   `json:"question"`
	Choices  []string `json:"choices"`
}

func NewQuestionSerializer(qIndex int, q entity.Question) QuestionSerializer {
	return QuestionSerializer{
		Index:    qIndex,
		Question: q.Question,
		Choices:  q.Answers,
	}
}

type GameInfoSerializer struct {
	CurrentQuestion   QuestionSerializer `json:"currentQuestion"`
	QuestionStartedAt int64              `json:"questionStartedAt"`
	QuestionEndsAt    int64              `json:"questionEndsAt"`
}

func NewGameInfoSerializer(lobby entity.Lobby) GameInfoSerializer {
	gameInfoSerialized := GameInfoSerializer{
		CurrentQuestion: NewQuestionSerializer(
			lobby.GameInfo.CurrentQuestion,
			lobby.Questions[lobby.GameInfo.CurrentQuestion],
		),
		QuestionStartedAt: lobby.GameInfo.CurrentQuestionStartedAt.Unix(),
		QuestionEndsAt:    lobby.GameInfo.CurrentQuestionEndsAt.Unix(),
	}

	return gameInfoSerialized
}

type LeaderboardItem struct {
	DisplayName string `json:"displayName"`
	Score       int64  `json:"score"`
}
type ResultSerializer struct {
	Winner      string            `json:"winner"`
	WinnerScore int64             `json:"winnerScore"`
	Leaderboard []LeaderboardItem `json:"leaderboard"`
}

func NewResultSerializer(lobby entity.Lobby) ResultSerializer {
	scores := lobby.Scores()
	return ResultSerializer{
		Winner:      scores[0].DisplayName,
		WinnerScore: scores[0].Score,
		Leaderboard: lo.Map(scores, func(item entity.Score, _ int) LeaderboardItem {
			return LeaderboardItem{
				DisplayName: item.DisplayName,
				Score:       item.Score,
			}
		}),
	}
}

type LobbySerializer struct {
	ID           string                          `json:"id"`
	State        string                          `json:"state"`
	Participants map[int64]ParticipantSerializer `json:"participants"`
	GameInfo     GameInfoSerializer              `json:"gameInfo"`
	Result       ResultSerializer                `json:"result"`
}

func NewLobbySerializer(lobby entity.Lobby) LobbySerializer {
	return LobbySerializer{
		ID:    lobby.ID,
		State: lobby.State,
		Participants: lo.MapValues[int64, entity.UserState, ParticipantSerializer](
			lobby.UserState,
			func(value entity.UserState, key int64) ParticipantSerializer {
				return NewParticipantSerializer(value, key, lobby.GameInfo.CorrectAnswers[key])
			},
		),
		GameInfo: NewGameInfoSerializer(lobby),
		Result:   NewResultSerializer(lobby),
	}
}

type EventSerializer struct {
	Type      string            `json:"type,omitempty"`
	AccountID int64             `json:"accountId,omitempty"`
	Account   AccountSerializer `json:"account,omitempty"`
}

func NewEventSerializer(info events.EventInfo) EventSerializer {
	return EventSerializer{
		Type:      info.Type.Type(),
		AccountID: info.AccountID,
		Account:   NewAccountSerializer(info.Account),
	}
}

type AccountSerializer struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
}

func NewAccountSerializer(account entity.Account) AccountSerializer {
	return AccountSerializer{
		ID:          account.ID,
		DisplayName: account.DisplayName,
	}
}

type FullAccountSerializer struct {
	entity.Account
}

func NewFullAccountSerializer(account entity.Account) FullAccountSerializer {
	return FullAccountSerializer{
		Account: account,
	}
}

type EventResponseSerializer struct {
	Lobby LobbySerializer `json:"lobby"`
	Event EventSerializer `json:"event,omitempty"`
	Hash  string          `json:"hash"`
}

func NewEventResponseSerializer(lobby entity.Lobby, info events.EventInfo, hash string) EventResponseSerializer {
	return EventResponseSerializer{
		Lobby: NewLobbySerializer(lobby),
		Event: NewEventSerializer(info),
		Hash:  hash,
	}
}

func Hash(t any) (string, error) {
	h, err := hashstructure.Hash(t, hashstructure.FormatV1, nil)
	if err != nil {
		logrus.WithError(err).WithField("item", t).Errorln("couldn't generate hash")
		return "", err
	}
	return strconv.FormatUint(h, 10), nil
}
