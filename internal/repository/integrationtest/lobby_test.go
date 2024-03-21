package integrationtest

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"kingscomp/internal/entity"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"testing"
)

func TestLobby_Ready(t *testing.T) {
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx := context.Background()
	lr := repository.NewLobbyRedisRepository(redisClient)

	err = lr.Save(ctx, entity.Lobby{
		ID:           "1",
		Participants: []int64{1, 2},
		UserState: map[int64]entity.UserState{
			1: {},
			2: {},
		},
	})

	assert.NoError(t, err)

	err = lr.UpdateUserState(ctx, "1", 1, "isReady", true)
	assert.NoError(t, err)

	lobby, err := lr.Get(ctx, entity.NewID("lobby", 1))
	assert.NoError(t, err)
	assert.True(t, lobby.UserState[1].IsReady)
}
