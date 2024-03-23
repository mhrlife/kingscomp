package events

import (
	"context"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
	"kingscomp/pkg/jsonhelper"
	"sync"
	"sync/atomic"
)

type Keys struct {
	keys map[string]*InMemoryEvents
	l    sync.Mutex
}

func NewKeys() *Keys {
	return &Keys{keys: make(map[string]*InMemoryEvents)}
}
func (k *Keys) Get(key string) *InMemoryEvents {
	k.l.Lock()
	defer k.l.Unlock()
	im, ok := k.keys[key]
	if ok {
		return im
	}
	k.keys[key] = NewInMemoryEvents()
	return k.keys[key]
}

type RedisPubSub struct {
	dedicated   rueidis.DedicatedClient
	client      rueidis.Client
	pattern     string
	closeClient func()

	cancelFunc context.CancelFunc
	ctx        context.Context
	isClosed   atomic.Bool

	keys *Keys
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
		keys:        NewKeys(),
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
	logrus.WithFields(logrus.Fields{
		"key":  key,
		"type": t.Type(),
	}).Info("dispatching a new message")
	return nil
}

func (r *RedisPubSub) Register(key string, t EventType, callback Callback) (func(), error) {
	inMem := r.keys.Get(key)
	clean, err := inMem.Register(t, callback)
	if err != nil {
		return nil, err
	}
	return func() {
		clean()
	}, nil
}

func (r *RedisPubSub) listen() {
	ch := make(chan struct{})
	go func() {
		wait := r.dedicated.SetPubSubHooks(rueidis.PubSubHooks{
			OnMessage: func(msg rueidis.PubSubMessage) {
				key := msg.Channel
				inMem := r.keys.Get(key)
				eventInfo := jsonhelper.Decode[EventInfo]([]byte(msg.Message))

				logrus.WithFields(logrus.Fields{
					"key":      msg.Channel,
					"type":     eventInfo.Type.Type(),
					"type_num": eventInfo.Type,
					"userId":   eventInfo.AccountID,
				}).Infoln("new pub/sub message recieved")

				if err := inMem.Dispatch(eventInfo.Type, eventInfo); err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"msg": msg,
						"key": key,
					}).Errorln("error while dispatching event")
				}
			},
			OnSubscription: func(s rueidis.PubSubSubscription) {
				ch <- struct{}{}
			},
		})
		r.dedicated.Do(r.ctx, r.client.B().Psubscribe().Pattern(r.pattern).Build())
		err := <-wait
		logrus.WithError(err).WithField("pattern", r.pattern).Errorln("PUB/SUB error")
		select {
		case <-r.ctx.Done():
			fmt.Println("ctx is done")
		default:
			fmt.Println("ctx is not done")

		}
	}()
	<-ch
	logrus.WithFields(logrus.Fields{
		"pattern": r.pattern,
	}).Info("subscription started")
}

func (r *RedisPubSub) Close() error {
	if r.cancelFunc != nil {
		r.cancelFunc()
	}
	r.closeClient()
	return nil
}

func (r *RedisPubSub) Clean(key string, t EventType) error {
	inMem := r.keys.Get(key)
	return inMem.Clean(t)
}
