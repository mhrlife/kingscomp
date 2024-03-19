package gameserver

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestEvents_DispatchAndRegister(t *testing.T) {
	e := NewEvents()
	var wg sync.WaitGroup
	wg.Add(2)
	c1 := e.Register(EventReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})
	c2 := e.Register(EventReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})

	e.Dispatch(EventReady, EventInfo{
		AccountID: 1,
	})
	wg.Wait()

	assert.Equal(t, 2, e.ListenerCount(EventReady))
	c1()
	assert.Equal(t, 1, e.ListenerCount(EventReady))
	c2()
	assert.Equal(t, 0, e.ListenerCount(EventReady))

	wg = sync.WaitGroup{}
	wg.Add(2)

	e.Register(EventReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})

	e.Register(EventReady, func(info EventInfo) {
		assert.Equal(t, info.AccountID, int64(1))
		wg.Done()
	})

	e.Dispatch(EventReady, EventInfo{
		AccountID: 1,
	})
	wg.Wait()

	assert.Equal(t, 2, e.ListenerCount(EventReady))
	e.Clean(EventReady)
	assert.Equal(t, 0, e.ListenerCount(EventReady))
}
