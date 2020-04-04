package eventlistener

import (
	"fmt"
)

type EventType int

type Event struct {
	Offset    uint64      `json:"offset"`
	Sender    string      `json:"sender"`
	CasinoID  string      `json:"casino_id"`
	GameID    string      `json:"game_id"`
	RequestID string      `json:"req_id"`
	EventType EventType   `json:"event_type"`
	Data      interface{} `json:"data"` // TODO: ??? interface or json raw message ?
}

type EventMessage struct {
	Offset uint64   `json:"offset"` // last event.offset
	Events []*Event `json:"events"`
}

func (e EventType) ToString() string {
	return fmt.Sprintf("event_%d", e)
}
