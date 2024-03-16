package integrationtest

import (
	"context"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"kingscomp/internal/entity"
	"kingscomp/internal/matchmaking"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"sync"
	"testing"
	"time"
)

func TestMatchmaking_Join(t *testing.T) {
	ctx := context.Background()
	timeout := time.Second * 10
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	mm := matchmaking.NewRedisMatchmaking(redisClient, repository.NewLobbyRedisRepository(redisClient))

	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, _, err := mm.Join(ctx, id, timeout)
			assert.NoError(t, err)
			assert.NotEqual(t, "", lobby)
			wg.Done()
		}()
	}

	testJoin(10)
	testJoin(11)
	testJoin(12)
	testJoin(13)

	<-time.After(time.Millisecond * 500)

	assert.Equal(t, int64(4), zCount(t, redisClient, "matchmaking"))

	lobby, _, err := mm.Join(ctx, 14, timeout)
	assert.NoError(t, err)
	assert.NotEqual(t, "", lobby.ID)
	wg.Wait()
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

func TestMatchmaking_JoinWithManyLobbies(t *testing.T) {
	ctx := context.Background()
	timeout := time.Second * 10
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	lobbyRepository := repository.NewLobbyRedisRepository(redisClient)
	accRepository := repository.NewAccountRedisRepository(redisClient)
	mm := matchmaking.NewRedisMatchmaking(redisClient, lobbyRepository)

	for i := 0; i < 100; i++ {
		accRepository.Save(context.Background(), entity.Account{
			ID:        int64(i),
			FirstName: "Mohammad",
		})
	}

	counter := NewCCounter[string]()
	uCounter := NewCCounter[int64]()
	var wg sync.WaitGroup
	testJoin := func(id int64) {
		wg.Add(1)
		go func() {
			lobby, created, err := mm.Join(ctx, id, timeout)
			assert.NoError(t, err)
			if created {
				counter.Incr(lobby.ID)
			}
			wg.Done()
		}()
	}

	s := time.Now()
	for i := 0; i < 1000; i++ {
		testJoin(int64(i) + 1)
	}

	wg.Wait()
	fmt.Println("Took", time.Since(s))

	assert.Len(t, counter.counter, 200)

	// each lobby must have only 5 users
	for lobbyId, _ := range counter.counter {
		lobby, err := lobbyRepository.Get(context.Background(), entity.NewID("lobby", lobbyId))
		assert.NoError(t, err)
		assert.Len(t, lobby.Participants, 5)
		for _, participant := range lobby.Participants {
			uCounter.Incr(participant)
		}
	}

	// each user must have joined only one lobby
	for _, count := range uCounter.counter {
		assert.Equal(t, 1, count)
	}

	// check whether account's current game lobby is correct
	acc, err := accRepository.Get(context.Background(), entity.NewID("account", 50))
	assert.NoError(t, err)
	assert.NotEqual(t, "", acc.CurrentLobby)
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
