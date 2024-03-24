package integrationtest

import (
	"context"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"kingscomp/internal/entity"
	"kingscomp/internal/matchmaking"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"sync"
	"testing"
	"time"
)

const maxLobbySize = 5

func TestMatchmaking(t *testing.T) {
	suite.Run(t, new(MatchmakingTestSuite))
}

type MatchmakingTestSuite struct {
	suite.Suite
	mm          matchmaking.Matchmaking
	ctx         context.Context
	timeout     time.Duration
	redisClient rueidis.Client

	lobby   repository.Lobby
	account repository.Account
}

func (s *MatchmakingTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.timeout = time.Second * 10
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(s.T(), err)
	qr := repository.NewQuestionRedisRepository(redisClient)
	ar := repository.NewAccountRedisRepository(redisClient)
	lr := repository.NewLobbyRedisRepository(redisClient)
	mm := matchmaking.NewRedisMatchmaking(
		redisClient,
		lr,
		qr,
		ar,
	)
	err = qr.PushActiveQuestion(s.ctx,
		entity.Question{ID: "1"},
		entity.Question{ID: "2"},
		entity.Question{ID: "3"},
		entity.Question{ID: "4"},
		entity.Question{ID: "5"},
		entity.Question{ID: "6"},
	)
	assert.NoError(s.T(), err)

	for i := 0; i < 100; i++ {
		ar.Save(context.Background(), entity.Account{
			ID:        int64(i),
			FirstName: "Mohammad",
		})
	}

	s.mm = mm
	s.redisClient = redisClient
	s.lobby = lr
	s.account = ar
}

func (s *MatchmakingTestSuite) TearDownSubTest() {
	flushAll(s.T(), s.redisClient)
}

func (s *MatchmakingTestSuite) TestMatchmaking_Join() {
	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, _, err := s.mm.Join(s.ctx, id, time.Second)
			assert.NoError(s.T(), err)
			assert.NotEqual(s.T(), "", lobby.ID)
			wg.Done()
		}()
	}

	for i := 0; i < maxLobbySize-1; i++ {
		testJoin(int64(3 + i))
	}

	<-time.After(time.Millisecond * 500)

	assert.Equal(s.T(), int64(maxLobbySize-1), zCount(s.T(), s.redisClient, "matchmaking"))

	lobby, _, err := s.mm.Join(s.ctx, 14, s.timeout)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), "", lobby.ID)
	wg.Wait()
}

func (s *MatchmakingTestSuite) TestMatchmaking_JoinTimeout() {
	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, _, err := s.mm.Join(s.ctx, id, time.Millisecond*100)
			assert.ErrorIs(s.T(), err, matchmaking.ErrTimeout)
			assert.Equal(s.T(), "", lobby.ID)
			wg.Done()
		}()
	}

	testJoin(10)

	<-time.After(time.Millisecond * 500)

	assert.Equal(s.T(), int64(0), zCount(s.T(), s.redisClient, "matchmaking"))
}

type CCounter[T comparable] struct {
	sync.Mutex
	counter map[T]int
}

func NewCCounter[T comparable]() CCounter[T] {
	return CCounter[T]{
		Mutex:   sync.Mutex{},
		counter: make(map[T]int),
	}
}

func (l *CCounter[T]) Incr(id T) {
	l.Lock()
	defer l.Unlock()

	l.counter[id]++
}

func (s *MatchmakingTestSuite) TestMatchmaking_JoinWithManyLobbies() {
	counter := NewCCounter[string]()
	uCounter := NewCCounter[int64]()
	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, _, err := s.mm.Join(s.ctx, id, s.timeout)
			assert.NoError(s.T(), err)
			counter.Incr(lobby.ID)
			wg.Done()
		}()
	}

	st := time.Now()
	for i := 0; i < maxLobbySize*1000; i++ {
		testJoin(int64(i) + 1)
	}

	wg.Wait()
	fmt.Println("Took", time.Since(st))

	assert.Len(s.T(), counter.counter, 1000)

	// each lobby must have only 5 users
	for lobbyId, count := range counter.counter {
		lobby, err := s.lobby.Get(context.Background(), entity.NewID("lobby", lobbyId))
		assert.NoError(s.T(), err)
		assert.Len(s.T(), lobby.Participants, maxLobbySize)
		assert.Equal(s.T(), count, maxLobbySize)
		for _, participant := range lobby.Participants {
			uCounter.Incr(participant)
		}
	}

	// each user must have joined only one lobby
	for _, count := range uCounter.counter {
		assert.Equal(s.T(), 1, count)
	}

	// check whether account's current game lobby is correct
	acc, err := s.account.Get(context.Background(), entity.NewID("account", 50))
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), "", acc.CurrentLobby)
}

func zCount(t *testing.T, redisClient rueidis.Client, key string) int64 {
	count, err := redisClient.Do(context.Background(),
		redisClient.B().Zcount().Key("matchmaking").Min("-inf").Max("+inf").Build(),
	).ToInt64()
	assert.NoError(t, err)
	return count
}

func keys(t *testing.T, redisClient rueidis.Client, pattern string) []string {
	keys, err := redisClient.Do(context.Background(),
		redisClient.B().Keys().Pattern(pattern).Build(),
	).AsStrSlice()
	assert.NoError(t, err)
	return keys
}

func flushAll(t *testing.T, c rueidis.Client) {
	assert.NoError(t, c.Do(context.Background(), c.B().Flushall().Build()).Error())
}
