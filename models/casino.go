package models

type CasinoMeta struct {
	ApiURL string `json:"apiUrl"`
}

type Casino struct {
	Id       uint64      `json:"id"`
	Contract string      `json:"contract"`
	Paused   bool        `json:"paused"`
	Meta     *CasinoMeta `json:"meta"`
}

type GameParam struct {
	Type  uint16 `json:"type"`
	Value uint64 `json:"value"`
}

type CasinoGame struct {
	Id     uint64      `json:"gameId"`
	Paused bool        `json:"paused"`
	Params []GameParam `json:"params"`
}
