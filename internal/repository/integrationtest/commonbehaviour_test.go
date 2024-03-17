package integrationtest

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"kingscomp/internal/entity"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"strconv"
	"testing"
)

type testType struct {
	ID   string
	Name string
}

func (t testType) EntityID() entity.ID {
	return entity.NewID("testType", t.ID)
}

func TestCommonBehaviourSetAndGet(t *testing.T) {
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx := context.Background()

	cb := repository.NewRedisCommonBehaviour[testType](redisClient)
	err = cb.Save(ctx, testType{
		ID:   "12",
		Name: "Sajad Jalilian",
	})
	assert.NoError(t, err)

	err = cb.Save(ctx, testType{
		ID:   "13",
		Name: "Amirreza",
	})
	assert.NoError(t, err)

	val, err := cb.Get(ctx, entity.NewID("testType", "12"))
	assert.NoError(t, err)
	assert.Equal(t, "Sajad Jalilian", val.Name)
	assert.Equal(t, "12", val.ID)

	val, err = cb.Get(ctx, entity.NewID("testType", "13"))
	assert.NoError(t, err)
	assert.Equal(t, "Amirreza", val.Name)
	assert.Equal(t, "13", val.ID)

	err = cb.Save(ctx, testType{
		ID:   "13",
		Name: "yasin",
	})

	assert.NoError(t, err)
	val, err = cb.Get(ctx, entity.NewID("testType", "13"))
	assert.NoError(t, err)
	assert.Equal(t, "yasin", val.Name)

	val, err = cb.Get(ctx, entity.NewID("testType", "14"))
	assert.ErrorIs(t, repository.ErrNotFound, err)
}

func TestCommonBehaviourMGet(t *testing.T) {
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx := context.Background()
	cb := repository.NewRedisCommonBehaviour[testType](redisClient)

	for i := 0; i < 10; i++ {
		err = cb.Save(ctx, testType{
			ID:   strconv.Itoa(i),
			Name: "Behzad",
		})
		assert.NoError(t, err)
	}

	items, err := cb.MGet(ctx,
		entity.NewID("testType", 2),
		entity.NewID("testType", 3),
		entity.NewID("testType", 4),
	)
	assert.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, "Behzad", items[0].Name)

}

func TestCommonBehaviourMGetNotExists(t *testing.T) {
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx := context.Background()
	cb := repository.NewRedisCommonBehaviour[testType](redisClient)

	items, err := cb.MGet(ctx,
		entity.NewID("testType", -100),
		entity.NewID("testType", -200),
		entity.NewID("testType", -300),
	)

	assert.NoError(t, err)
	assert.Len(t, items, 0)

}
