package models

type Casino struct {
	Id       uint64 `json:"id"`
	Contract string `json:"contract"`
	Paused   bool   `json:"paused"`
}
