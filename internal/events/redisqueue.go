package events

import (
	"context"
	"errors"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
	"kingscomp/pkg/jsonhelper"
	"time"
)

type RedisQueue struct {
	rdb rueidis.Client

	ctx    context.Context
	cancel context.CancelFunc
	key    string
	inMem  *InMemoryEvents
}

func NewRedisQueue(ctx context.Context, key string, rdb rueidis.Client) *RedisQueue {
	ctx, cancel := context.WithCancel(ctx)
	rq := &RedisQueue{
		rdb:    rdb,
		ctx:    ctx,
		cancel: cancel,
		key:    key,
		inMem:  NewInMemoryEvents(),
	}
	rq.listen()
	return rq
}

func (r *RedisQueue) Dispatch(ctx context.Context, t EventType, info EventInfo) error {
	info.Type = t
	cmd := r.rdb.B().Rpush().Key(r.key).Element(string(jsonhelper.Encode(info))).Build()
	err := r.rdb.Do(ctx, cmd).Error()
	if err != nil {
		logrus.WithError(err).Errorln("couldn't enqueue in redis")
		return err
	}
	logrus.WithFields(logrus.Fields{
		"key":  r.key,
		"type": t.Type(),
	}).Info("enqueue in redis completed")
	return nil
}

func (r *RedisQueue) Register(t EventType, callback Callback) (func(), error) {
	clean, err := r.inMem.Register(t, callback)
	if err != nil {
		return nil, err
	}
	return func() {
		clean()
	}, nil
}

func (r *RedisQueue) listen() {
	go func() {
		for {
			cmd := r.rdb.B().Blpop().Key(r.key).Timeout(0).Build()
			val, err := r.rdb.Do(r.ctx, cmd).AsStrSlice()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				logrus.WithError(err).WithField("key", r.key).Errorln("something bad happened while poping")
				<-time.After(time.Second)
				continue
			}

			if len(val) == 0 {
				logrus.WithError(err).WithField("key", r.key).Errorln("this part shouldn't be executed")
				<-time.After(time.Second)
				continue
			}
			eventInfo := jsonhelper.Decode[EventInfo]([]byte(val[1]))
			logrus.WithError(err).
				WithField("key", r.key).
				WithField("type", eventInfo.Type.Type()).
				Info("popped a new item from queue")
			if err := r.inMem.Dispatch(eventInfo.Type, eventInfo); err != nil {
				logrus.WithError(err).WithField("key", r.key).Errorln("error while dispatching queue item")
			}
		}
	}()
}
