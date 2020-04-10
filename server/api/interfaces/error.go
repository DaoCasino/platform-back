package interfaces

type WsError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}
