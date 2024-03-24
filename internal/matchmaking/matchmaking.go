package matchmaking

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/config"
	"kingscomp/internal/entity"
	"kingscomp/internal/repository"
	"kingscomp/pkg/jsonhelper"
	"kingscomp/pkg/randhelper"
	"strconv"
	"time"
)

var (
	ErrBadRedisResponse = errors.New("bad redis response")
	ErrTimeout          = errors.New("lobby queue timeout")
)

//go:embed matchmaking.lua
var matchmakingScript string

type Matchmaking interface {
	Join(ctx context.Context, userId int64, timeout time.Duration) (entity.Lobby, bool, error)
	Leave(ctx context.Context, userId int64) error
}

var _ Matchmaking = &RedisMatchmaking{}

type RedisMatchmaking struct {
	client            rueidis.Client
	matchMakingScript *rueidis.Lua
	lobby             repository.Lobby
	account           repository.Account
	question          repository.Question
}

func NewRedisMatchmaking(client rueidis.Client,
	lobby repository.Lobby,
	question repository.Question,
	account repository.Account,
) *RedisMatchmaking {
	script := rueidis.NewLuaScript(matchmakingScript)
	return &RedisMatchmaking{
		client:            client,
		matchMakingScript: script,
		lobby:             lobby,
		question:          question,
		account:           account,
	}
}

func (r RedisMatchmaking) Join(ctx context.Context, userId int64, timeout time.Duration) (entity.Lobby, bool, error) {

	defer r.Leave(context.Background(), userId)

	resp, err := r.matchMakingScript.Exec(ctx, r.client,
		[]string{"matchmaking", "matchmaking"},
		[]string{fmt.Sprint(config.Default.LobbyMaxPlayer - 1),
			strconv.FormatInt(time.Now().Add(-time.Minute*2).Unix(), 10),
			uuid.New().String(), strconv.FormatInt(userId, 10),
			strconv.FormatInt(time.Now().Unix(), 10),
		}).ToArray()
	if err != nil {
		logrus.WithError(err).Errorln("couldn't join the match making")
		return entity.Lobby{}, false, err
	}

	// inside a queue, we must listen to the pub/sub
	if len(resp) == 1 {
		logrus.WithField("userId", userId).Info("waiting for a lobby")
		cmd := r.client.B().Brpop().
			Key(fmt.Sprintf("matchmaking:%d", userId)).Timeout(timeout.Seconds()).Build()
		result, err := r.client.Do(ctx, cmd).AsStrSlice()
		if err != nil {
			if errors.Is(err, rueidis.Nil) {
				return entity.Lobby{}, false, ErrTimeout
			}
			logrus.WithError(err).Errorln("couldn't get matchmaking notice from redis")
			return entity.Lobby{}, false, err
		}

		if len(result) < 2 {
			return entity.Lobby{}, false, ErrTimeout
		}

		logrus.WithFields(logrus.Fields{
			"userId": userId,
			"lobby":  result[1],
		}).Info("found a new lobby")
		lobby, err := r.lobby.Get(ctx, entity.NewID("lobby", result[1]))
		return lobby, false, err
	}

	// you have just created a lobby
	if len(resp) == 3 {
		lobbyId, _ := resp[1].ToString()
		logrus.WithField("lobbyId", lobbyId).Info("created a new lobby")
		matchedUsers, _ := resp[2].AsIntSlice()

		lobby, err := r.createNewLobby(ctx, lobbyId, matchedUsers)
		if err != nil {
			return entity.Lobby{}, false, err
		}

		return lobby, true, err
	}

	return entity.Lobby{}, false, ErrBadRedisResponse
}

func (r RedisMatchmaking) Leave(ctx context.Context, userId int64) error {
	cmds := make([]rueidis.Completed, 0)
	cmds = append(cmds,
		r.client.B().Zrem().Key("matchmaking").Member(strconv.FormatInt(userId, 10)).Build(),
		r.client.B().Del().Key(fmt.Sprintf("matchmaking:%d", userId)).Build(),
	)
	if err := repository.ReduceRedisResponseError(r.client.DoMulti(ctx, cmds...)); err != nil {
		logrus.WithError(err).Errorln("couldn't successfully leave the match making")
		return err
	}
	return nil
}

func (r RedisMatchmaking) createNewLobby(ctx context.Context, lobbyId string, users []int64) (entity.Lobby, error) {
	cmds := make([]rueidis.Completed, 0)

	// get lobby questions
	activeQuestionsCount, err := r.question.GetActiveQuestionsCount(ctx)
	if err != nil {
		return entity.Lobby{}, err
	}

	questionIndexes := randhelper.GenerateDistinctRandomNumbers(config.Default.LobbyQuestionCount, 0, activeQuestionsCount-1)
	questions, err := r.question.GetActiveQuestions(ctx, questionIndexes...)
	if err != nil {
		return entity.Lobby{}, err
	}

	fullUsers, err := r.account.MGet(ctx, lo.Map(users, func(item int64, index int) entity.ID {
		return entity.NewID("account", item)
	})...)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't fetch accounts to add them to lobby")
		return entity.Lobby{}, err
	}
	// create the lobby
	userStates := make(map[int64]entity.UserState, len(users))
	for _, user := range users {
		userStates[user] = entity.UserState{}
	}
	for _, user := range fullUsers {
		s := userStates[user.ID]
		s.DisplayName = user.DisplayName
		userStates[user.ID] = s
	}
	lobby := entity.Lobby{
		ID:            lobbyId,
		Participants:  users,
		CreatedAtUnix: time.Now().Unix(),
		Questions:     questions,
		UserState:     userStates,
		State:         "created",
	}

	cmds = append(cmds,
		r.client.B().JsonSet().Key(entity.NewID("lobby", lobby.ID).String()).Path(".").Value(
			string(jsonhelper.Encode(lobby)),
		).Build(),
		r.client.B().Expire().Key(entity.NewID("lobby", lobby.ID).String()).Seconds(
			int64((time.Minute * 15).Seconds()),
		).Build(),
	)

	// update participants current lobby
	for _, participant := range users {
		userMatchmakingListKey := fmt.Sprintf("matchmaking:%d", participant)
		cmds = append(cmds,
			r.client.B().JsonSet().
				Key(entity.NewID("account", participant).String()).Path("$..current_lobby").
				Value(fmt.Sprintf(`"%s"`, lobbyId)).Xx().Build(),
			r.client.B().Rpush().Key(userMatchmakingListKey).Element(lobbyId).Build(),
			r.client.B().Expire().Key(userMatchmakingListKey).Seconds(10).Build(),
		)
	}

	resp := r.client.DoMulti(ctx, cmds...)
	err = repository.ReduceRedisResponseError(resp, rueidis.Nil)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't create the matchmaking lobby")
		return entity.Lobby{}, err
	}
	return lobby, nil
}
