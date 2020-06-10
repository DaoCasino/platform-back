package models

type GameMeta struct {
	ManifestURL string `json:"manifestURL"`
}

type Game struct {
	Id        uint64    `json:"id"`
	Contract  string    `json:"contract"`
	ParamsCnt uint16    `json:"paramsCnt"`
	Paused    int       `json:"paused"`
	Meta      *GameMeta `json:"meta"`
}
