package eventlistener

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventType_ToString(t *testing.T) {
	var e EventType = 1
	assert.Equal(t, "event_1", e.ToString())
}
