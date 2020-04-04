package eventlistener

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	actionMonitorAddr = ":8888"
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
	defer cancel()

	listener := NewEventListener(actionMonitorAddr, nil)
	if err := listener.ListenAndServe(parentContext); err != nil {
		t.Skip("listen error", err.Error())
		return
	}

	ok, err := listener.Subscribe(0, 0)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestEventListener_Unsubscribe(t *testing.T) {
	parentContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener := NewEventListener(actionMonitorAddr, nil)
	if err := listener.ListenAndServe(parentContext); err != nil {
		t.Skip("listen error", err.Error())
		return
	}

	// Unsubscribing from a topic that is not subscribed to
	ok, err := listener.Unsubscribe(666)
	require.Error(t, err)
	assert.False(t, ok)

	ok, err = listener.Subscribe(0, 0)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = listener.Unsubscribe(0)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestEventListener_EventsMessage(t *testing.T) {
	parentContext, cancel := context.WithCancel(context.Background())
	events := make(chan *EventMessage)
	defer func() {
		close(events)
		cancel()
	}()

	listener := NewEventListener(":8888", events)
	if err := listener.ListenAndServe(parentContext); err != nil {
		t.Skip("listen error", err.Error())
		return
	}

	ok, err := listener.Subscribe(0, 0)
	require.NoError(t, err)
	assert.True(t, ok)

	waitContext, cancelWait := context.WithTimeout(parentContext, 2*time.Second)

loop:
	for {
		select {
		case <-waitContext.Done():
			t.Log("no events")
			break loop
		case event := <-events:
			if len(event.Events) == 0 {
				t.Error("received 0 events; want more")
				break loop
			}
			t.Logf("%+v", event.Events[0])
			break loop

		}
	}

	cancelWait()
}
