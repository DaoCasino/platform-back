package session

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"platform-backend/models"
	"platform-backend/server/api"
)

type Manager interface {
	NewConnection(wsConn *websocket.Conn, wsApi *api.WsApi)
	HasSessionByUser(userName string) bool
	AuthUser(uid uuid.UUID, user *models.User) error
}