package webapp

import (
	"github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/entity"
	"kingscomp/internal/gameserver"
	"strconv"
)

type ParticipantSerializer struct {
	ID                      int64  `json:"id"`
	DisplayName             string `json:"displayName"`
	IsReady                 bool   `json:"isReady"`
	IsResigned              bool   `json:"isResigned"`
	CurrentQuestionAnswered bool   `json:"currentQuestionAnswered"`
}

func NewParticipantSerializer(state entity.UserState, id int64, currentQuestion int) ParticipantSerializer {
	return ParticipantSerializer{
		ID:                      id,
		DisplayName:             state.DisplayName,
		IsReady:                 state.IsReady,
		IsResigned:              state.IsResigned,
		CurrentQuestionAnswered: state.LastAnsweredQuestionIndex >= currentQuestion,
	}
}

type LobbySerializer struct {
	ID           string                          `json:"id"`
	State        string                          `json:"state"`
	Participants map[int64]ParticipantSerializer `json:"participants"`
}

func NewLobbySerializer(lobby entity.Lobby) LobbySerializer {
	return LobbySerializer{
		ID:    lobby.ID,
		State: lobby.State,
		Participants: lo.MapValues[int64, entity.UserState, ParticipantSerializer](
			lobby.UserState,
			func(value entity.UserState, key int64) ParticipantSerializer {
				return NewParticipantSerializer(value, key, lobby.GameInfo.CurrentQuestion)
			},
		),
	}
}

type EventSerializer struct {
	Type      string            `json:"type,omitempty"`
	AccountID int64             `json:"accountId,omitempty"`
	Account   AccountSerializer `json:"account,omitempty"`
}

func NewEventSerializer(info gameserver.EventInfo) EventSerializer {
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

func NewEventResponseSerializer(lobby entity.Lobby, info gameserver.EventInfo, hash string) EventResponseSerializer {
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
