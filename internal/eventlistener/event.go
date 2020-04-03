package eventlistener

import (
	"encoding/json"
)

type Event struct {
	Offset    uint64          `json:"offset"`
	Sender    string          `json:"sender"`
	CasinoID  string          `json:"casino_id"`
	GameID    string          `json:"game_id"`
	RequestID string          `json:"req_id"`
	EventType int             `json:"event_type"`
	Data      json.RawMessage `json:"data"`
}

func parseEvent(data []byte) (*Event, error) {
	fields := new(Event)
	if err := json.Unmarshal(data, fields); err != nil {
		return nil, err
	}

	return fields, nil
}
