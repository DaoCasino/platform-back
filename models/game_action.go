package models

type GameAction struct {
	Type   uint16   `json:"type"`
	Params []uint64 `json:"params"`
}
