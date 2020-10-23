package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type GameSessionUpdateType uint16

type GameSessionUpdateMsg struct {
	SessionID  string                `json:"sessionId"`
	UpdateType GameSessionUpdateType `json:"updateType"`
	Timestamp  time.Time             `json:"timestamp"`
	Data       json.RawMessage       `json:"data"`
	Offset     *uint64               `json:"offset"`
}

func ToGameSessionUpdateMsg(u *GameSessionUpdate) *GameSessionUpdateMsg {
	return &GameSessionUpdateMsg{
		SessionID:  strconv.FormatUint(u.SessionID, 10),
		UpdateType: u.UpdateType,
		Timestamp:  u.Timestamp,
		Data:       u.Data,
		Offset:     u.Offset,
	}
}

type GameSessionUpdate struct {
	SessionID  uint64                `json:"sessionId"`
	UpdateType GameSessionUpdateType `json:"updateType"`
	Timestamp  time.Time             `json:"timestamp"`
	Data       json.RawMessage       `json:"data"`
	Offset     *uint64               `json:"offset"`
}

const (
	SessionCreatedUpdate GameSessionUpdateType = iota
	SessionStartedUpdate
	GameActionRequestedUpdate
	GameMessageUpdate
	GameFinishedUpdate
	GameFailedUpdate
)
