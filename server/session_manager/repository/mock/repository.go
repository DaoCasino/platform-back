package mock

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"platform-backend/models"
	"platform-backend/server/api"
)

type MockRepository struct {
	mock.Mock
}

func (r *MockRepository) AddSession(context context.Context, wsConn *websocket.Conn, wsApi *api.WsApi) {
	r.Called(wsConn, wsApi)
}

func (r *MockRepository) HasSessionByUser(accountName string) bool {
	args := r.Called(accountName)

	return args.Get(0).(bool)
}

func (r *MockRepository) SetUser(uid uuid.UUID, user *models.User) error {
	args := r.Called(uid, user)

	return args.Error(0)
}
