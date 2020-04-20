package models

type Game struct {
	Id           uint64 `json:"id"`
	Contract     string `json:"contract"`
	ParamsCnt    uint16 `json:"params_cnt"`
	Paused       int    `json:"paused"`
}
