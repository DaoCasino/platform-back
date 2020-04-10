package session

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"platform-backend/models"
	"platform-backend/server/api"
	"sync"
)

type ManagerImpl struct {
	sync.Mutex
	// main sessions registry
	sessionById   map[uuid.UUID]*Session
	// session by account name map
	sessionByUser map[string]*Session
}

func (m *ManagerImpl) removeSession(uid uuid.UUID) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.sessionById[uid]; !ok {
		// session doesn't exists
		return
	}

	// remove from by user map
	if m.sessionById[uid].user != nil {
		delete(m.sessionByUser, m.sessionById[uid].user.AccountName)
	}

	// remove from main map
	delete(m.sessionById, uid)
}

func NewSessionManager() *ManagerImpl {
	manager := new(ManagerImpl)
	manager.sessionById = make(map[uuid.UUID]*Session)
	manager.sessionByUser = make(map[string]*Session)

	return manager
}

func (m *ManagerImpl) NewConnection(wsConn *websocket.Conn, wsApi *api.WsApi) {
	m.Lock()

	var session *Session

	defer func() { // prevent potential dead-lock
		m.Unlock()
		session.Run()
	}()

	session = NewSession(context.Background(), wsConn, wsApi, func() {
		m.removeSession(session.uuid)
	})

	m.sessionById[session.uuid] = session

	log.Debug().Msgf("New session started, uid: %m", session.uuid.String())
}

func (m *ManagerImpl) HasSessionByUser(userName string) bool {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.sessionByUser[userName]; ok {
		return true
	}
	return false
}

func (m *ManagerImpl) AuthUser(uid uuid.UUID, user *models.User) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.sessionById[uid]; !ok {
		return errors.New("session not found")
	}

	// set user info
	m.sessionById[uid].user = user

	return nil
}