package models

import (
	"encoding/json"
	"time"
)

type GameSessionUpdateType uint16

type GameSessionUpdate struct {
	SessionID  uint64                `json:"sessionId"`
	UpdateType GameSessionUpdateType `json:"updateType"`
	Timestamp  time.Time             `json:"timestamp"`
	Data       json.RawMessage       `json:"data"`
}

const (
	SessionCreatedUpdate GameSessionUpdateType = iota
	SessionStartedUpdate
	GameActionRequestedUpdate
	GameMessageUpdate
	GameFinishedUpdate
	GameFailedUpdate
)
