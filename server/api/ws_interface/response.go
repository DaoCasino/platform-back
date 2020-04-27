package ws_interface

type WsResponse struct {
	Type    string      `json:"type"`
	Id      string      `json:"id"`
	Status  string      `json:"status"`
	Payload interface{} `json:"payload"`
}
