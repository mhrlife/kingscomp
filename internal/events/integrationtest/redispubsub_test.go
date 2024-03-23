package integrationtest

import (
	"context"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"kingscomp/internal/events"
	"kingscomp/internal/repository/redis"
	"testing"
	"time"
)

type RedisPubSubSuite struct {
	suite.Suite
	ps *events.RedisPubSub
	rc rueidis.Client

	ctx    context.Context
	cancel context.CancelFunc
	key    string
}

func TestRedisPubSub(t *testing.T) {
	rc, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	key := "lobby:1"
	s := &RedisPubSubSuite{
		ps:     events.NewRedisPubSub(ctx, rc, "lobby.*"),
		cancel: cancel,
		key:    key,
		ctx:    ctx,
		rc:     rc,
	}
	suite.Run(t, s)
}

func (r *RedisPubSubSuite) TearDownSuite() {
	r.ps.Close()
	<-time.After(time.Millisecond * 100)
}

func (r *RedisPubSubSuite) TestPubSub() {
	ch := make(chan struct{})
	cancel, _ := r.ps.Register("lobby.1", events.EventAny, func(info events.EventInfo) {
		assert.Equal(r.T(), events.EventUserAnswer, info.Type)
		ch <- struct{}{}
	})
	defer cancel()
	err := r.ps.Dispatch(r.ctx, "lobby.1", events.EventUserAnswer, events.EventInfo{})
	assert.NoError(r.T(), err)
	select {
	case <-ch:
	case <-time.After(time.Second):
		r.T().Fatal("timeout message")
	}
}

func (r *RedisPubSubSuite) TestPubSubClose() {
	cancel, _ := r.ps.Register("lobby.1", events.EventAny, func(info events.EventInfo) {
		r.T().Fatal("this block must not run")
	})
	cancel()

	ch := make(chan struct{})
	c2, _ := r.ps.Register("lobby.1", events.EventAny, func(info events.EventInfo) {
		ch <- struct{}{}
	})
	defer c2()
	err := r.ps.Dispatch(r.ctx, "lobby.1", events.EventUserAnswer, events.EventInfo{})
	assert.NoError(r.T(), err)
	select {
	case <-ch:
	case <-time.After(time.Millisecond * 100):
		r.T().Fatal("timeout")
	}
}
