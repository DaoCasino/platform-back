package models

import "encoding/json"

type GameMeta struct {
	ManifestURL string `json:"manifestURL"`
	Ext json.RawMessage `json:"ext"`
}

type Game struct {
	Id        uint64    `json:"id"`
	Contract  string    `json:"contract"`
	ParamsCnt uint16    `json:"paramsCnt"`
	Paused    int       `json:"paused"`
	Meta      *GameMeta `json:"meta"`
}
