package events

import (
	"context"
	"errors"
	"fmt"
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

var _ PubSub = &RedisPubSub{}

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
	err := r.client.Do(ctx, cmd).Error()
	if err != nil {
		logrus.WithError(err).Errorln("couldn't dispatch pub/sub message")
		return err
	}
	return nil
}

func (r *RedisPubSub) Register(key string, t EventType, callback Callback) (func(), error) {
	iInMem, _ := r.keys.LoadOrStore(key, NewInMemoryEvents())
	inMem := iInMem.(*InMemoryEvents)
	clean, err := inMem.Register(t, callback)
	fmt.Println(t.Type(), inMem)
	if err != nil {
		return nil, err
	}
	return func() {
		clean()
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

			logrus.WithFields(logrus.Fields{
				"key":      msg.Channel,
				"type":     eventInfo.Type.Type(),
				"type_num": eventInfo.Type,
				"userId":   eventInfo.AccountID,
			}).Infoln("new pub/sub message recieved")

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

func (r *RedisPubSub) Clean(key string, t EventType) error {
	iInMem, ok := r.keys.Load(key)
	if !ok {
		return nil
	}
	inMem := iInMem.(*InMemoryEvents)
	return inMem.Clean(t)
}
