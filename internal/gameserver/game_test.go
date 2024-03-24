package gameserver

import (
	"context"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/maps"
	"kingscomp/internal/entity"
	"kingscomp/internal/events"
	"kingscomp/internal/repository"
	"kingscomp/internal/service"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type KVRepository[T entity.Entity] struct {
	kvStore map[entity.ID]T
	l       sync.RWMutex
}

func NewKVRepository[T entity.Entity]() *KVRepository[T] {
	return &KVRepository[T]{kvStore: make(map[entity.ID]T)}
}

func (kv *KVRepository[T]) Get(ctx context.Context, id entity.ID) (T, error) {
	kv.l.Lock()
	defer kv.l.Unlock()
	t, ok := kv.kvStore[id]
	if !ok {
		return t, repository.ErrNotFound
	}
	return t, nil
}

func (kv *KVRepository[T]) getLobby(id string) T {
	l, _ := kv.Get(context.Background(), entity.NewID("lobby", id))
	return l
}

func (kv *KVRepository[T]) clear() {
	kv.l.Lock()
	defer kv.l.Unlock()

	kv.kvStore = make(map[entity.ID]T)
}

func (kv *KVRepository[T]) Save(ctx context.Context, ent T) error {
	kv.l.Lock()
	defer kv.l.Unlock()
	kv.kvStore[ent.EntityID()] = ent
	return nil
}

func (kv *KVRepository[T]) MGet(ctx context.Context, ids ...entity.ID) ([]T, error) {
	var items []T
	for _, id := range ids {
		t, err := kv.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, nil
}

func (kv *KVRepository[T]) MSet(ctx context.Context, ents ...T) error {
	for _, e := range ents {
		kv.Save(ctx, e)
	}
	return nil
}

func (kv *KVRepository[T]) SetField(ctx context.Context, id entity.ID, fieldName string, value any) error {
	kv.l.Lock()
	defer kv.l.Unlock()
	return nil
}

func (kv *KVRepository[T]) AllIDs(ctx context.Context, prefix string) ([]string, error) {
	kv.l.Lock()
	defer kv.l.Unlock()
	return lo.Map(maps.Keys(kv.kvStore), func(item entity.ID, _ int) string {
		return item.String()
	}), nil
}

type LobbyKVRepository struct {
	*KVRepository[entity.Lobby]
}

func NewLobbyKVRepository() *LobbyKVRepository {
	return &LobbyKVRepository{KVRepository: NewKVRepository[entity.Lobby]()}
}

func (l *LobbyKVRepository) UpdateUserState(ctx context.Context,
	lobbyId string, userId int64, key string, val any) error {
	us := l.kvStore[entity.NewID("lobby", lobbyId)].UserState[userId]
	switch key {
	case "isResigned":
		us.IsResigned = val.(bool)
	case "isReady":
		us.IsReady = val.(bool)
	case "lastAnsweredQuestionIndex":
		us.LastAnsweredQuestionIndex = val.(int)
	}
	l.kvStore[entity.NewID("lobby", lobbyId)].UserState[userId] = us
	return nil
}

type GameServerTestSuite struct {
	suite.Suite
	gs *GameServer

	accounts  *KVRepository[entity.Account]
	lobbies   *LobbyKVRepository
	lobbyId   string
	pubSubKey string
	game      *Game
}

func TestGameServer(t *testing.T) {
	accounts := NewKVRepository[entity.Account]()
	lobbies := NewLobbyKVRepository()

	suite.Run(t, &GameServerTestSuite{
		accounts:  accounts,
		lobbies:   lobbies,
		lobbyId:   "test-lobby-1",
		pubSubKey: "lobby.test-lobby-1",
	})
}

func testQuestion(id string, correctAnswer int) entity.Question {
	return entity.Question{
		ID:            id,
		Question:      id + "-q",
		Answers:       []string{"a1", "a2", "a3", "a4"},
		CorrectAnswer: correctAnswer,
	}
}
func (s *GameServerTestSuite) SetupTest() {
	if s.gs != nil && s.gs.cancelCtx != nil {
		s.gs.cancelCtx()
	}

	s.accounts.clear()
	s.lobbies.clear()

	s.accounts.Save(context.Background(), entity.Account{
		ID:           1,
		FirstName:    "User1",
		CurrentLobby: s.lobbyId,
	})
	s.accounts.Save(context.Background(), entity.Account{
		ID:           2,
		FirstName:    "User2",
		CurrentLobby: s.lobbyId,
	})
	s.lobbies.Save(context.Background(), entity.Lobby{
		ID:            s.lobbyId,
		Participants:  []int64{1, 2},
		CreatedAtUnix: time.Now().Unix(),
		Questions: []entity.Question{
			testQuestion("q1", 1),
			testQuestion("q2", 2),
			testQuestion("q3", 3),
			testQuestion("q4", 1),
		},
		UserState: map[int64]entity.UserState{
			1: {
				DisplayName: "User1",
			},
			2: {
				DisplayName: "User2",
			},
		},
		State: "created",
	})
	s.gs = NewGameServer(&service.App{
		Account: service.NewAccountService(s.accounts),
		Lobby:   service.NewLobbyService(s.lobbies),
	}, events.NewInMemPubSub(), Config{
		ReminderToReadyAfter: time.Millisecond * 50,
		ReadyDeadline:        time.Millisecond * 100,
		QuestionTimeout:      time.Millisecond * 100,
		GetReadyDuration:     time.Millisecond * 100,
		LobbyAge:             time.Minute,
	})
	s.game = s.gs.MustGame(s.lobbyId)
}

func (s *GameServerTestSuite) TestJoinNotification() {
	<-time.After(time.Millisecond)
	assert.Equal(s.T(), "created", s.game.lobby.State)
	// the first user connects
	s.ready(1)
	<-time.After(time.Millisecond)
	assert.Equal(s.T(), "created", s.game.lobby.State)
	// the second user connects
	s.ready(2)
	<-time.After(time.Millisecond)
	assert.NotEqual(s.T(), "created", s.game.lobby.State)
}

func (s *GameServerTestSuite) TestJoinResignedNotification() {
	<-time.After(time.Millisecond)
	assert.Equal(s.T(), "created", s.game.lobby.State)
	// the first user connects
	s.ready(1)
	<-time.After(time.Millisecond)
	assert.Equal(s.T(), "created", s.game.lobby.State)
	// the second user connects
	s.resign(2)
	<-time.After(time.Millisecond)
	assert.NotEqual(s.T(), "created", s.game.lobby.State)
}

func (s *GameServerTestSuite) TestTimeoutResignedNotification() {
	s.ready(1)
	eCounter := int32(0)
	c, _ := s.game.Events.Register(s.pubSubKey, events.EventLateResign, func(info events.EventInfo) {
		atomic.AddInt32(&eCounter, 1)
	})
	defer c()
	<-time.After(time.Millisecond * 150)
	assert.Equal(s.T(), int32(1), eCounter)
	assert.NotEqual(s.T(), "created", s.game.lobby.State)
}

func (s *GameServerTestSuite) TestAnswerQuestionAnswer() {
	s.ready(1)
	s.ready(2)
	<-time.After(time.Millisecond * 210)
	assert.Equal(s.T(), "started", s.game.lobby.State)
	assert.Equal(s.T(), 0, s.game.lobby.GameInfo.CurrentQuestion)
	// this is a bad question index, must be 1
	s.game.Events.Dispatch(context.Background(), s.pubSubKey, events.EventUserAnswer, events.EventInfo{
		AccountID:     1,
		QuestionIndex: 1,
		UserAnswer:    1,
	})
	<-time.After(time.Millisecond)
	assert.Len(s.T(), s.game.lobby.GameInfo.CorrectAnswers, 0)
	// this is a correct question index
	s.game.Events.Dispatch(context.Background(), s.pubSubKey, events.EventUserAnswer, events.EventInfo{
		AccountID:     1,
		QuestionIndex: 0,
		UserAnswer:    1,
	})
	<-time.After(time.Millisecond * 5)
	assert.Len(s.T(), s.game.lobby.GameInfo.CorrectAnswers, 1)
	assert.True(s.T(), s.game.lobby.GameInfo.CorrectAnswers[1][0].Correct)
	// user 2 didn't answer, check if it has automatically got false
	<-time.After(time.Millisecond * 100)
	assert.Len(s.T(), s.game.lobby.GameInfo.CorrectAnswers[2], 1)
	assert.False(s.T(), s.game.lobby.GameInfo.CorrectAnswers[2][0].Correct)
	// both didn't answer the second question
	<-time.After(time.Millisecond * 100)
	assert.Equal(s.T(), 2, s.game.lobby.GameInfo.CurrentQuestion)
	assert.Len(s.T(), s.game.lobby.GameInfo.CorrectAnswers[1], 2)
	assert.Len(s.T(), s.game.lobby.GameInfo.CorrectAnswers[2], 2)
	assert.False(s.T(), s.game.lobby.GameInfo.CorrectAnswers[1][1].Correct)
	assert.False(s.T(), s.game.lobby.GameInfo.CorrectAnswers[2][1].Correct)
}

func (s *GameServerTestSuite) TestLobbyEnded() {
	s.ready(1)
	s.ready(2)
	<-time.After(time.Millisecond * 210)
	assert.Equal(s.T(), "started", s.game.lobby.State)

	for i := 0; i < len(s.game.lobby.Questions); i++ {
		assert.Equal(s.T(), i, s.game.lobby.GameInfo.CurrentQuestion)
		s.answer(i, 1, 1)
		s.answer(i, 1, 2)
		<-time.After(time.Millisecond * 10)
		if i < len(s.game.lobby.Questions)-1 {
			assert.Equal(s.T(), i+1, s.game.lobby.GameInfo.CurrentQuestion)
		} else {
			assert.Equal(s.T(), i, s.game.lobby.GameInfo.CurrentQuestion)
		}
	}

	<-time.After(time.Millisecond * 10)
	assert.Equal(s.T(), "ended", s.game.lobby.State)
	<-time.After(time.Second)

}

func (s *GameServerTestSuite) answer(question, answer int, accountId int64) {
	s.game.Events.Dispatch(context.Background(), s.pubSubKey, events.EventUserAnswer, events.EventInfo{
		AccountID:     accountId,
		QuestionIndex: question,
		UserAnswer:    answer,
	})
}

func (s *GameServerTestSuite) ready(userId int64) {
	s.lobbies.UpdateUserState(context.Background(), s.lobbyId, userId, "isReady", true)
	s.game.Events.Dispatch(context.Background(), s.pubSubKey, events.EventUserReady, events.EventInfo{AccountID: userId})
}

func (s *GameServerTestSuite) resign(userId int64) {
	s.lobbies.UpdateUserState(context.Background(), s.lobbyId, userId, "isResigned", true)
	s.game.Events.Dispatch(context.Background(), s.pubSubKey, events.EventUserResigned, events.EventInfo{AccountID: userId})
}
