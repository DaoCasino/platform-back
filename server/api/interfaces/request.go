package interfaces

import (
	"platform-backend/models"
	"platform-backend/usecases"
)

type WsRequest struct {
	Request string      `json:"request"`
	Id      string      `json:"id"`
	Payload interface{} `json:"payload"`
}

type ApiRequest struct {
	UseCases *usecases.UseCases
	User     *models.User
	Message  []byte
}
