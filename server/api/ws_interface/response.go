package ws_interface

type WsResponse struct {
	Type    string      `json:"type"`
	Id      string      `json:"id"`
	Status  string      `json:"status"`
	Payload interface{} `json:"payload"`
}

type WsUpdate struct {
	Type    string      `json:"type"`
	Reason  string      `json:"reason"`
	Time    int64       `json:"time"`
	Payload interface{} `json:"payload"`
}
