package events

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestEvents_DispatchAndRegister(t *testing.T) {
	e := NewInMemoryEvents()
	var wg sync.WaitGroup
	wg.Add(3)
	c1, _ := e.Register(EventUserReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})
	c2, _ := e.Register(EventUserReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})
	c3, _ := e.Register(EventAny, func(info EventInfo) {
		assert.Equal(t, info.Type, EventUserReady)
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})

	e.Dispatch(EventUserReady, EventInfo{
		AccountID: 1,
	})
	wg.Wait()

	assert.Equal(t, 2, e.listenerCount(EventUserReady))
	c1()
	assert.Equal(t, 1, e.listenerCount(EventUserReady))
	c2()
	assert.Equal(t, 0, e.listenerCount(EventUserReady))
	c3()
	assert.Equal(t, 0, e.listenerCount(EventAny))

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

	assert.Equal(t, 2, e.listenerCount(EventUserReady))
	e.Clean(EventUserReady)
	assert.Equal(t, 0, e.listenerCount(EventUserReady))
}
