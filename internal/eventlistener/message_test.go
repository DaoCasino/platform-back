package eventlistener

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEventListener_sendRequest(t *testing.T) {
	listener := &EventListener{
		send: make(chan *responseQueue),
	}

	rawResult := json.RawMessage("true")

	go func() {
		responseQueue := <-listener.send
		responseQueue.response <- &responseMessage{ID: &responseQueue.ID, Result: rawResult}
		close(responseQueue.response)
	}()

	request := newRequestMessage("test", nil)
	response, err := listener.sendRequest(request)
	require.NoError(t, err)
	assert.Equal(t, rawResult, response.Result)

	close(listener.send)
}
