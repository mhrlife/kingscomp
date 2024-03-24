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

type RedisStreamSuite struct {
	suite.Suite
	rc rueidis.Client
	rq *events.RedisQueue

	ctx    context.Context
	cancel context.CancelFunc
}

func TestRedisStream(t *testing.T) {
	rc, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	s := &RedisStreamSuite{
		cancel: cancel,
		ctx:    ctx,
		rc:     rc,
	}
	suite.Run(t, s)
}

func (r *RedisStreamSuite) TearDownSuite() {
	r.cancel()
	<-time.After(time.Millisecond * 100)
}

func (r *RedisStreamSuite) BeforeTest(suiteName, testName string) {
	r.cancel()
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.rq = events.NewRedisQueue(r.ctx, "key", r.rc)
}

func (r *RedisStreamSuite) TestQueueSimple() {
	ch := make(chan struct{})
	cancel, err := r.rq.Register(events.EventAny, func(info events.EventInfo) {
		ch <- struct{}{}
	})
	defer cancel()
	r.NoError(err)
	r.NoError(r.rq.Dispatch(r.ctx, events.EventLateResign, events.EventInfo{}))
	select {
	case <-ch:
	case <-time.After(time.Millisecond * 100):
		r.Fail("timeout")

	}
}

func (r *RedisStreamSuite) TestQueueCanceled() {
	ch := make(chan struct{})
	cancel, err := r.rq.Register(events.EventAny, func(info events.EventInfo) {
		ch <- struct{}{}
	})
	r.NoError(err)
	cancel()
	r.NoError(r.rq.Dispatch(r.ctx, events.EventLateResign, events.EventInfo{}))
	select {
	case <-ch:
		r.Fail("timeout")
	case <-time.After(time.Millisecond * 100):

	}
}
