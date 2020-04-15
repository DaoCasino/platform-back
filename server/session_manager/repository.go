package session_manager

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"platform-backend/models"
	"platform-backend/server/api"
)

type Repository interface {
	AddSession(context context.Context, wsConn *websocket.Conn, wsApi *api.WsApi)
	HasSessionByUser(accountName string) bool
	SetUser(uid uuid.UUID, user *models.User) error
}