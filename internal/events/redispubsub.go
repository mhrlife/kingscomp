package events

import (
	"context"
	"errors"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
	"kingscomp/pkg/jsonhelper"
	"sync"
	"sync/atomic"
)

type RedisPubSub struct {
	dedicated   rueidis.DedicatedClient
	client      rueidis.Client
	pattern     string
	closeClient func()

	cancelFunc context.CancelFunc
	ctx        context.Context
	isClosed   atomic.Bool

	keys sync.Map
}

//var _ PubSub = &RedisPubSub{}

func NewRedisPubSub(ctx context.Context, client rueidis.Client, pattern string) *RedisPubSub {
	ctx, cancelFunc := context.WithCancel(ctx)
	dedicated, closeFunc := client.Dedicate()
	r := &RedisPubSub{
		client:      client,
		dedicated:   dedicated,
		pattern:     pattern,
		closeClient: closeFunc,
		ctx:         ctx,
		cancelFunc:  cancelFunc,
	}
	go r.listen()
	return r
}

func (r *RedisPubSub) Dispatch(ctx context.Context, key string, t EventType, info EventInfo) error {
	info.Type = t
	cmd := r.client.B().Publish().Channel(key).Message(
		string(jsonhelper.Encode(info)),
	).Build()
	return r.client.Do(ctx, cmd).Error()
}

func (r *RedisPubSub) Register(key string, t EventType, callback Callback) (func(), error) {
	iInMem, _ := r.keys.LoadOrStore(key, NewInMemoryEvents())
	inMem := iInMem.(*InMemoryEvents)
	clean, err := inMem.Register(t, callback)
	if err != nil {
		return nil, err
	}
	return func() {
		l := clean()
		if l == 0 {
			r.keys.Delete(key)
		}
	}, nil
}

func (r *RedisPubSub) listen() {
	go func() {
		cmd := r.dedicated.B().Psubscribe().Pattern(r.pattern).Build()
		err := r.dedicated.Receive(r.ctx, cmd, func(msg rueidis.PubSubMessage) {
			key := msg.Channel
			iInMem, ok := r.keys.Load(key)
			if !ok {
				return
			}
			eventInfo := jsonhelper.Decode[EventInfo]([]byte(msg.Message))
			if err := iInMem.(*InMemoryEvents).Dispatch(eventInfo.Type, eventInfo); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"msg": msg,
					"key": key,
				}).Errorln("error while dispatching event")
			}
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			logrus.WithError(err).WithField("pattern", r.pattern).Errorln("subscriber error")
		}
	}()
}

func (r *RedisPubSub) Close() error {
	if r.cancelFunc != nil {
		r.cancelFunc()
	}
	r.closeClient()
	return nil
}
