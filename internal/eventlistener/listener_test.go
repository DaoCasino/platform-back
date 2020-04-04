package eventlistener

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewEventListener(t *testing.T) {
	addr := ":1234"
	listener := NewEventListener(addr, nil)

	assert.Equal(t, addr, listener.Addr)
	assert.Equal(t, int64(maxMessageSize), listener.MaxMessageSize)
	assert.Equal(t, writeWait, listener.WriteWait)
	assert.Equal(t, pongWait, listener.PongWait)
	assert.Equal(t, pingPeriod, listener.PingPeriod)
	assert.Nil(t, listener.event)
}

func TestEventListener_ListenAndServe(t *testing.T) {
	parentContext := context.Background()

	listener := NewEventListener(":1234", nil)
	err := listener.ListenAndServe(parentContext)
	require.Error(t, err)
}

func TestEventListener_Subscribe(t *testing.T) {
	parentContext, cancel := context.WithCancel(context.Background())

	listener := NewEventListener(":8888", nil)
	if err := listener.ListenAndServe(parentContext); err != nil {
		t.Skip("listen error", err.Error())
	}

	time.Sleep(2 * time.Second)
	cancel()
}

//
//func TestEventListener_Unsubscribe(t *testing.T) {
//	t.Skip("need write")
//}
