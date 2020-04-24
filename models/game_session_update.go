package models

import (
	"encoding/json"
	"time"
)

type GameSessionUpdate struct {
	SessionID  uint64          `json:"sessionId"`
	UpdateType uint16          `json:"updateType"`
	Timestamp  time.Time       `json:"timestamp"`
	Data       json.RawMessage `json:"data"`
}
