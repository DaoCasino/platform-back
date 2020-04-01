package models

import "time"

type GameSessionUpdate struct {
	SessionID  uint64
	UpdateType uint16
	Timestamp  time.Time
	Data       []byte
}
