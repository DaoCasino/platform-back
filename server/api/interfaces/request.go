package interfaces

import (
	"encoding/json"
	"platform-backend/models"
	"platform-backend/usecases"
)

type WsRequest struct {
	Request string      `json:"request"`
	Id      string      `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

type ApiRequest struct {
	UseCases *usecases.UseCases
	User     *models.User
	Data     *WsRequest
}
