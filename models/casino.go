package models

type Casino struct {
	Id       uint64 `json:"id"`
	Contract string `json:"contract"`
	Paused   bool   `json:"paused"`
}

type GameParam struct {
	Type  uint16 `json:"type"`
	Value uint64 `json:"value"`
}

type CasinoGame struct {
	Id     uint64      `json:"game_id"`
	Paused bool        `json:"paused"`
	Params []GameParam `json:"params"`
}
