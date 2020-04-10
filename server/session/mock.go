package session

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"platform-backend/models"
	"platform-backend/server/api"
)

type ManagerMock struct {
	mock.Mock
}

func (m *ManagerMock) NewConnection(wsConn *websocket.Conn, wsApi *api.WsApi) {
}

func (m *ManagerMock) HasSessionByUser(userName string) bool {
	args := m.Called(userName)

	return args.Get(0).(bool)
}

func (m *ManagerMock) AuthUser(uid uuid.UUID, user *models.User) error {
	args := m.Called(uid, user)

	return args.Error(0)
}
