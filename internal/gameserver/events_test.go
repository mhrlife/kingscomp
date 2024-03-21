package gameserver

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestEvents_DispatchAndRegister(t *testing.T) {
	e := NewEvents()
	var wg sync.WaitGroup
	wg.Add(3)
	c1 := e.Register(EventUserReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})
	c2 := e.Register(EventUserReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})
	c3 := e.Register(EventAny, func(info EventInfo) {
		assert.Equal(t, info.Type, EventUserReady)
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})

	e.Dispatch(EventUserReady, EventInfo{
		AccountID: 1,
	})
	wg.Wait()

	assert.Equal(t, 2, e.ListenerCount(EventUserReady))
	c1()
	assert.Equal(t, 1, e.ListenerCount(EventUserReady))
	c2()
	assert.Equal(t, 0, e.ListenerCount(EventUserReady))
	c3()
	assert.Equal(t, 0, e.ListenerCount(EventAny))

	wg = sync.WaitGroup{}
	wg.Add(2)

	e.Register(EventUserReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})

	e.Register(EventUserReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})

	e.Dispatch(EventUserReady, EventInfo{
		AccountID: 1,
	})
	wg.Wait()

	assert.Equal(t, 2, e.ListenerCount(EventUserReady))
	e.Clean(EventUserReady)
	assert.Equal(t, 0, e.ListenerCount(EventUserReady))
}
