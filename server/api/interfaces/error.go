package interfaces

type WsError struct {
	Code    uint64  `json:"code"`
	Message string `json:"message"`
}
